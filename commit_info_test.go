package gitbackend

import (
	"testing"
	"time"
)

func TestNewCommitInfo(t *testing.T) {
	subject := NewCommitInfo("Paul", "p@example.com", "Have fun.",
		time.Date(2014, 10, 17, 13, 37, 0, 0, &time.Location{}))
	if subject == nil {
		t.Fatal("Expected *CommitInfo, got nil")
	}
	if subject.AuthorName() != "Paul" {
		t.Fatalf("Expected 'Paul', actual %v", subject.AuthorName())
	}
	if subject.AuthorEmail() != "p@example.com" {
		t.Fatalf("Expected 'p@example.com', actual %v", subject.AuthorEmail())
	}
	if subject.Message() != "Have fun." {
		t.Fatalf("Expected 'Have fun.', actual %v", subject.Message())
	}
	if subject.Time().String() != "2014-10-17 13:37:00 +0000 UTC" {
		t.Fatalf("Expected '2014-10-17 13:37:00 +0000 UTC', actual %v", subject.Time())
	}
}
