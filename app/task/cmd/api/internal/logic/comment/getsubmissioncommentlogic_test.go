package comment

import (
	"MuXiFresh-Be-2.0/app/task/cmd/api/internal/types"
	"testing"
)

const zeroCommentID = "000000000000000000000000"

func TestBuildCommentTreeNoPanicWithoutRoot(t *testing.T) {
	comments := []types.Comment{
		{CommentID: "c1", FatherID: "c0"},
		{CommentID: "c2", FatherID: "c1"},
	}

	tree := buildCommentTree(comments)

	if len(tree) != 0 {
		t.Fatalf("expected empty tree, got %d roots", len(tree))
	}
}

func TestBuildCommentTreeSkipsOrphanReply(t *testing.T) {
	comments := []types.Comment{
		{CommentID: "root1", FatherID: zeroCommentID},
		{CommentID: "root2", FatherID: zeroCommentID},
		{CommentID: "reply1", FatherID: "root1"},
		{CommentID: "reply2", FatherID: "missing"},
	}

	tree := buildCommentTree(comments)

	if len(tree) != 2 {
		t.Fatalf("expected 2 roots, got %d", len(tree))
	}
	if len(tree[0].Replies) != 1 || tree[0].Replies[0].CommentID != "reply1" {
		t.Fatalf("expected reply1 under root1, got %+v", tree[0].Replies)
	}
	if len(tree[1].Replies) != 0 {
		t.Fatalf("expected root2 has no replies, got %+v", tree[1].Replies)
	}
}

func TestBuildCommentTreeNestedReplies(t *testing.T) {
	comments := []types.Comment{
		{CommentID: "root1", FatherID: zeroCommentID},
		{CommentID: "lvl1", FatherID: "root1"},
		{CommentID: "lvl2", FatherID: "lvl1"},
	}

	tree := buildCommentTree(comments)

	if len(tree) != 1 {
		t.Fatalf("expected 1 root, got %d", len(tree))
	}
	if len(tree[0].Replies) != 1 || tree[0].Replies[0].CommentID != "lvl1" {
		t.Fatalf("expected lvl1 under root1, got %+v", tree[0].Replies)
	}
	if len(tree[0].Replies[0].Replies) != 1 || tree[0].Replies[0].Replies[0].CommentID != "lvl2" {
		t.Fatalf("expected lvl2 under lvl1, got %+v", tree[0].Replies[0].Replies)
	}
}
