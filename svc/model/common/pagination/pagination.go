package pagination

import (
	"strconv"

	spb "github.com/q3k/bugless/proto/svc"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// The pagination package implements helpers for value-based pagination.
// Value-based pagination (in contrast to offset-based pagination) centers
// centers around pages selected by start value, instead of offset.
//
// For example, for a list of blog posts, traditional offset-based pagination
// would, in terms of a possible view, look something like this:
//
//   GET /posts?offset=10&count= 5
//
// This would in turn translate to a SQL query:
//
//   SELECT ... FROM posts
//   ORDER BY id
//   OFFSET 10
//   LIMIT 5
//
// This, however, is slow for the database - as the query results in having to
// prepare an entire 'view' of all posts up until the end of this page, and
// then discarding everything before the OFFSET value. Thus, the further you
// go down a list of posts, the slower the queries will be.
//
// What value-based pagination (sometimes called cursor-based pagination, or
// keyset pagination) does instead is to paginate by one of the values of the
// result. For example, the previous example would look like this:
//
//   GET /posts?after=123&count= 5
//
// This would in turn translate to a SQL query:
//
//   SELECT ... FROM posts
//   ORDER BY id
//   WHERE id > 123
//   LIMIT 5
//
// Notice that we replaced an OFFSET clause with a WHERE clause. This
// dramatically shortens execution time, especially if the field used in the
// WHERE clause is indexed.
//
// This approach, however, has two downsides:
//  - ordering by more than one field gets tricky, and if your database does
//    not support ordering of composite values your queries will get quite
//    complex.
//  - while going forwards through is easy, going backwards is not - some
//    storage of 'after' values per page is required on the client side, which
//    then also gets tricky with invalidations, and in general pushes quite a
//    bit of logic down to the consumer of this pagination API.
//
// Reference: Pagination done the Right Way, Markus Winand, 2013
// https://www.slideshare.net/slideshow/embed_code/22210863
//
// There are two levels where this pagination is present in Bugless:
//  - on gRPC APIs that return a list of results
//  - internally in the crdb backend to allow receiving smaller chunks of data
//    from CockroachDB
//
// This package implements helper glue that allows implementers of paginated,
// streaming APIs to send multiple queries to a backing store, and stream the
// resulting chunks to the API consumer. It manages clamping chunk size, checks
// exit conditions, and allows the API producer to include extra fields in the
// first returned chunk. It assumes that both the API consumer, and backing
// store use a value-based pagination system, and basically performs
// resampling of these to ensure database requests stay somewhat sane.
//
// To use this API, the producer needs to implement a ChunkSender function,
// then call Resample (or ResampleInt64 in case pagination values are int64s).

// V is the opaque, 'generic' type that's used to pass pagination 'values'
// around.
type V interface{}

// ChunkSender is a user declared function that retrieves a chunk of requested
// items from a backing store, sends the result to the requester, and reports
// on the number of items sent.
// This will be called repeatedly by the Paginator until the request (expressed
// in parameters to Resample) is fulfilled, or the backing store has ran out of
// data.
// The first time this function will be called, 'first' will be set to true. On
// all other calls, it will be false. start is the start value at which the
// backing store should retrieve date, and count is the amount of tlements that
// should be returned to the user. The function returns the amount of items it
// got from the backing store and sent to the user, the new start value at
// which further backing store requests should be made, and an error if one
// occured during execution.
type ChunkSender func(first bool, start V, count int64) (n int, ns V, err error)

// Resample calls ChunkSender to fullfil a request for (start, count) paginated
// data from an API consumer.
func Resample(start V, count int64, c ChunkSender) error {
	// Maximum request count sent to the backing store.
	maxChunkSize := int64(100)

	var sent int
	first := false
	for {
		// If we already sent what the consumer requested, we're done.
		if int64(sent) >= count {
			return nil
		}

		// Cap the backing store request count.
		chunkSize := count - int64(sent)
		if chunkSize > maxChunkSize {
			chunkSize = maxChunkSize
		}

		// Perform a request to the backing store by the producer-defined
		// ChunkSender.
		n, newStart, err := c(first, start, chunkSize)
		if err != nil {
			return err
		}

		// If the ChunkSender returned less data than expected, it means
		// there's no more data to request - we're done.
		if int64(n) < chunkSize {
			return nil
		}

		if first {
			first = false
		}
		sent += n
		start = newStart
	}
}

// ResampleInt64 runs Resample after parsing pagintion data from a proto,
// interpreting the start value as an int64.
func ResampleInt64(p *spb.PaginationSelector, c ChunkSender) error {
	var after int64
	var count int64
	var err error
	if p != nil {
		if p.After != "" {
			after, err = strconv.ParseInt(p.After, 10, 64)
			if err != nil {
				return status.Error(codes.InvalidArgument, "invalid pagination 'after'")
			}
		}
		count = p.Count
	}
	if count <= 0 {
		count = 100
	}
	return Resample(after, count, c)
}
