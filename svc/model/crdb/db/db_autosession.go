package db

import "context"

// autoSession is a db.Session, but wraps every call to child methods within
// a new session. This is useful for one-shot operations and for tests.
type autoSession struct {
	db  *database
	ctx context.Context
}

func (a *autoSession) Category() CategoryGetter {
	return &autoSessionCategory{a}
}

func (a *autoSession) Issue() IssueGetter {
	return &autoSessionIssue{a}
}

func (a *autoSession) Commit() error {
	panic("autoSession (from db.Database.Do) cannot be commited!")
}

func (a *autoSession) Rollback() error {
	panic("autoSession (from db.Database.Do) cannot be rolled back!")
}

type autoSessionCategory struct {
	*autoSession
}

type autoSessionIssue struct {
	*autoSession
}

// All praise Rob “Commander” Pike!
// (I'm sure there's a better way to do this)

func (c *autoSessionCategory) Get(uuid string) (*Category, error) {
	s := c.db.Begin(c.ctx)
	res, err := s.Category().Get(uuid)
	if err != nil {
		s.Rollback()
		return nil, err
	}
	return res, s.Commit()
}

func (c *autoSessionCategory) GetTree(uuid string, levels uint) (*CategoryNode, error) {
	s := c.db.Begin(c.ctx)
	res, err := s.Category().GetTree(uuid, levels)
	if err != nil {
		s.Rollback()
		return nil, err
	}
	return res, s.Commit()
}

func (c *autoSessionCategory) New(new *Category) (*Category, error) {
	s := c.db.Begin(c.ctx)
	res, err := s.Category().New(new)
	if err != nil {
		s.Rollback()
		return nil, err
	}
	return res, s.Commit()
}

func (c *autoSessionCategory) Update(cat *Category) error {
	s := c.db.Begin(c.ctx)
	err := s.Category().Update(cat)
	if err != nil {
		s.Rollback()
		return err
	}
	return s.Commit()
}

func (c *autoSessionCategory) Delete(uuid string) error {
	s := c.db.Begin(c.ctx)
	err := s.Category().Delete(uuid)
	if err != nil {
		s.Rollback()
		return err
	}
	return s.Commit()
}

func (c *autoSessionIssue) Get(id int64) (*Issue, error) {
	s := c.db.Begin(c.ctx)
	issue, err := s.Issue().Get(id)
	if err != nil {
		s.Rollback()
		return nil, err
	}
	return issue, s.Commit()
}

func (c *autoSessionIssue) Filter(filter IssueFilter, order IssueOrderBy, opts *IssueFilterOpts) ([]*Issue, error) {
	s := c.db.Begin(c.ctx)
	issues, err := s.Issue().Filter(filter, order, opts)
	if err != nil {
		s.Rollback()
		return nil, err
	}
	return issues, s.Commit()
}

func (c *autoSessionIssue) GetHistory(id int64, opts *IssueGetHistoryOpts) ([]*IssueUpdate, error) {
	s := c.db.Begin(c.ctx)
	issue, err := s.Issue().GetHistory(id, opts)
	if err != nil {
		s.Rollback()
		return nil, err
	}
	return issue, s.Commit()
}

func (c *autoSessionIssue) New(new *Issue) (*Issue, error) {
	s := c.db.Begin(c.ctx)
	issue, err := s.Issue().New(new)
	if err != nil {
		s.Rollback()
		return nil, err
	}
	return issue, s.Commit()
}

func (c *autoSessionIssue) Update(update *IssueUpdate) error {
	s := c.db.Begin(c.ctx)
	err := s.Issue().Update(update)
	if err != nil {
		s.Rollback()
		return err
	}
	return s.Commit()
}
