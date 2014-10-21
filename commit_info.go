package gitbackend

import (
  "time"
)

type CommitInfo struct {
  authorName, authorEmail, message string
  time                             time.Time
}

func (c *CommitInfo) AuthorName() string {
  return c.authorName
}

func (c *CommitInfo) AuthorEmail() string {
  return c.authorEmail
}

func (c *CommitInfo) Time() time.Time {
  return c.time
}

func (c *CommitInfo) Message() string {
  return c.message
}
