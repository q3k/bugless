// Copyright 2020 Sergiusz Bazanski <q3k@q3k.org>
// SPDX-License-Identifier: AGPL-3.0-or-later

package db

import (
	"context"

	"github.com/inconshreveable/log15"
	"github.com/lib/pq"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ErrorConverter struct {
	pgerrcodes map[string]error
}

func NewErrorConverter() *ErrorConverter {
	return &ErrorConverter{
		pgerrcodes: make(map[string]error),
	}
}

func (c *ErrorConverter) WithSyntaxError(err error) *ErrorConverter {
	c.pgerrcodes["42601"] = err
	return c
}

func (c *ErrorConverter) WithForeignKeyViolation(err error) *ErrorConverter {
	c.pgerrcodes["23503"] = err
	return c
}

func (c *ErrorConverter) WithUniqueConstraintViolation(err error) *ErrorConverter {
	c.pgerrcodes["23505"] = err
	return c
}

func (c *ErrorConverter) Convert(err error) error {
	if err == nil {
		return nil
	}

	// Directly return context errors.
	switch err {
	case context.Canceled:
		fallthrough
	case context.DeadlineExceeded:
		return err
	}

	switch uerr := err.(type) {
	// Parse database errors.
	case *pq.Error:
		code := string(uerr.Code)
		pretty, ok := c.pgerrcodes[code]
		if !ok {
			log15.Error("Unhandled postgres error", "err", err, "code", code)
			return status.Error(codes.Unavailable, "database error")
		}
		return pretty

	// Directly return gRPC errors.
	case interface{ GRPCStatus() *status.Status }:
		return uerr.GRPCStatus().Err()
	}

	log15.Error("Unhandled database error", "err", err)
	return status.Error(codes.Unavailable, "internal error")
}
