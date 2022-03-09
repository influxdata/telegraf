package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateCommitGroups(t *testing.T) {
	fixCommit1 := Commit{
		Type:    "fix",
		Subject: "new change",
	}
	fixCommit2 := Commit{
		Type:    "fix",
		Subject: "another change",
	}
	featCommit := Commit{
		Type: "feat",
	}

	commits := []*Commit{
		&fixCommit1,
		&featCommit,
		&fixCommit2,
	}

	commitGroups := CreateCommitGroups(commits)

	expectedCommitGroups := []CommitGroup{
		{
			Title: fixGroupTitle,
			Commits: []*Commit{
				&fixCommit2,
				&fixCommit1,
			},
		},
		{
			Title: featureGroupTitle,
			Commits: []*Commit{
				&featCommit,
			},
		},
	}

	require.Equal(t, commitGroups, expectedCommitGroups)
}
