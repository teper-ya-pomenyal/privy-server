package redis

import "testing"

func TestSelectSessionVictims(t *testing.T) {
	// id — это sessionID(token); считаем честно, чтобы тест ловил расхождение хэша
	tA, tB, tC := "token-a", "token-b", "token-c"
	idA, idB, idC := sessionID(tA), sessionID(tB), sessionID(tC)

	cases := []struct {
		name     string
		members  []string
		keep     string
		remove   []string
		expected []string
	}{
		{
			name:     "revoke all but keep",
			members:  []string{tA, tB, tC},
			keep:     idB,
			expected: []string{tA, tC},
		},
		{
			name:     "revoke listed only",
			members:  []string{tA, tB, tC},
			remove:   []string{idA, idC},
			expected: []string{tA, tC},
		},
		{
			name:     "keep wins over remove empty",
			members:  []string{tA},
			keep:     idA,
			expected: nil,
		},
		{
			name:     "unknown ids remove nothing",
			members:  []string{tA},
			remove:   []string{"deadbeef"},
			expected: nil,
		},
		{
			name:     "empty keep revokes everything",
			members:  []string{tA, tB},
			expected: []string{tA, tB},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			removeSet := make(map[string]struct{}, len(tc.remove))
			for _, id := range tc.remove {
				removeSet[id] = struct{}{}
			}
			got := selectSessionVictims(tc.members, tc.keep, removeSet)
			if len(got) != len(tc.expected) {
				t.Fatalf("victims = %v, expected %v", got, tc.expected)
			}
			set := make(map[string]bool, len(got))
			for _, v := range got {
				set[v] = true
			}
			for _, e := range tc.expected {
				if !set[e] {
					t.Fatalf("victims = %v, expected %v", got, tc.expected)
				}
			}
		})
	}
}

func TestSessionIDIsHexSha256(t *testing.T) {
	id := sessionID("some-refresh-token")
	if len(id) != 64 {
		t.Fatalf("session id length = %d, expected 64 hex chars", len(id))
	}
	if id != sessionID("some-refresh-token") {
		t.Fatal("session id must be deterministic")
	}
}
