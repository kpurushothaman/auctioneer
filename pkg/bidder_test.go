package resolver

import (
	"errors"
	"slices"
	"testing"
)

func lowBidder(bidTime string) Bidder {
	return Bidder{
		GUID: "Low",
		BidStrategy: BidStrategy{
			MaxBidCents:       2000,
			BidIncrementCents: 500,
		},
		InitialBid: Bid{
			AmountCents: 1000,
			CreatedAt:   justTime(bidTime),
		},
	}
}

func namedBidder(name, time string) Bidder {
	return Bidder{
		GUID: name,
		BidStrategy: BidStrategy{
			MaxBidCents:       30000,
			BidIncrementCents: 500,
		},
		InitialBid: Bid{
			AmountCents: 3000,
			CreatedAt:   justTime(time),
		},
	}
}

func namedBidderInitialAmount(name, time string, initialAmount int64) Bidder {
	return Bidder{
		GUID: name,
		BidStrategy: BidStrategy{
			MaxBidCents:       30000,
			BidIncrementCents: 500,
		},
		InitialBid: Bid{
			AmountCents: initialAmount,
			CreatedAt:   justTime(time),
		},
	}
}

func namedBidderInitialAmountMaxBid(name, time string, initialAmount, maxBid int64) Bidder {
	return Bidder{
		GUID: name,
		BidStrategy: BidStrategy{
			MaxBidCents:       maxBid,
			BidIncrementCents: 500,
		},
		InitialBid: Bid{
			AmountCents: initialAmount,
			CreatedAt:   justTime(time),
		},
	}
}

func highBidder(bidTime string) Bidder {
	return Bidder{
		GUID: "High",
		BidStrategy: BidStrategy{
			MaxBidCents:       30000,
			BidIncrementCents: 500,
		},
	}
}

func TestSortBidders(t *testing.T) {
	tests := map[string]struct {
		bidders  []Bidder
		expected []string
	}{
		"asc": {
			bidders: []Bidder{
				namedBidder("before_a", "2000-01-01 01:00:00"),
				namedBidder("a", "2007-01-01 00:00:00"),
				namedBidder("b", "2006-01-01 01:00:00"),
				namedBidder("c", "2001-01-01 01:00:00"),
			},
			expected: []string{"before_a", "c", "b", "a"},
		},
	}

	for name, tt := range tests {
		test := tt
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			sortBiddersByEarliest(test.bidders)
			for idx, bidder := range test.bidders {
				if test.expected[idx] != bidder.GUID {
					t.Fatalf("sorting bidder test failed, returned %s, expected %s", bidder.GUID, test.expected[idx])
				}
			}
		})
	}
}

func TestCompareBidders(t *testing.T) {
	tests := map[string]struct {
		a        Bidder
		b        Bidder
		expected int
	}{
		"a<b": {
			a:        lowBidder("2006-01-01 01:00:00"),
			b:        lowBidder("2007-01-01 01:00:00"),
			expected: -1,
		},
		"a>b": {
			a:        lowBidder("2106-01-01 01:00:00"),
			b:        lowBidder("2007-01-01 01:00:00"),
			expected: 1,
		},
		"a==b": {
			a:        lowBidder("2006-01-01 01:00:00"),
			b:        lowBidder("2006-01-01 01:00:00"),
			expected: 0,
		},
	}

	for name, tt := range tests {
		test := tt
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			actual := compareBiddersByBidTime(test.a, test.b)

			if actual != test.expected {
				t.Fatalf("compareBidders failed, returned %d, expected %d", actual, test.expected)
			}
		})
	}
}

func TestGetInitialBids(t *testing.T) {
	sorted := []Bidder{
		namedBidderInitialAmount("before_a", "2000-01-01 01:00:00", 1000),
		namedBidderInitialAmount("a", "2007-01-01 00:00:00", 1030),
		namedBidderInitialAmount("b", "2006-01-01 01:00:00", 5000),
		namedBidderInitialAmount("c", "2001-01-01 01:00:00", 833100),
	}

	sortBiddersByEarliest(sorted)

	tests := map[string]struct {
		bidders               []Bidder
		expected              map[string]int64
		expectedHighestBidder string
		expectedErr           error
	}{
		"non-error": {
			bidders: sorted,
			expected: map[string]int64{
				"before_a": 1000,
				"a":        1030,
				"b":        5000,
				"c":        833100,
			},
			expectedHighestBidder: "c",
			expectedErr:           nil,
		},
		"error": {
			bidders: []Bidder{
				namedBidderInitialAmount("a", "2007-01-01 00:00:00", 1030),
				namedBidderInitialAmount("b", "2006-01-01 01:00:00", 5000),
			},
			expected:    map[string]int64{},
			expectedErr: errors.New("getInitialBids(): passed array not sorted"),
		},
	}

	for name, tt := range tests {
		test := tt
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			initialBids, highestBidder, err := getInitialBids(test.bidders)

			for _, bidder := range test.bidders {
				if err != nil {
					if test.expectedErr.Error() != err.Error() {
						t.Fatalf("error mismatch, expected %s got %s", test.expectedErr, err)
					}
				}

				if test.expectedHighestBidder != highestBidder {
					t.Fatalf("GetInitialBids() test fail: highest expected %s, got %s", test.expectedHighestBidder, highestBidder)
				}

				if test.expected[bidder.GUID] != initialBids[bidder.GUID] {
					t.Fatalf("GetInitialBids() test fail:bidder %s, returned %d, expected %d", bidder.GUID, test.expected[bidder.GUID], initialBids[bidder.GUID])
				}
			}
		})
	}
}

func TestRemoveBiddersUnderMin(t *testing.T) {
	bidders := []Bidder{
		namedBidderInitialAmountMaxBid("before_a", "2000-01-01 01:00:00", 1000, 1499),
		namedBidderInitialAmountMaxBid("a", "2007-01-01 00:00:00", 1030, 2000),
		namedBidderInitialAmountMaxBid("b", "2006-01-01 01:00:00", 5000, 6000),
		namedBidderInitialAmountMaxBid("c", "2001-01-01 01:00:00", 833100, 9000000),
	}

	tests := map[string]struct {
		bidders  []Bidder
		min      int64
		expected []Bidder
	}{
		"happy path": {
			bidders: bidders,
			min:     2990,
			expected: []Bidder{
				namedBidderInitialAmountMaxBid("b", "2006-01-01 01:00:00", 5000, 6000),
				namedBidderInitialAmountMaxBid("c", "2001-01-01 01:00:00", 833100, 9000000),
			},
		},
	}

	for name, tt := range tests {
		test := tt
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			actual := removeBiddersUnderMin(test.bidders, test.min)

			for _, actualBidder := range actual {
				if !slices.Contains(test.expected, actualBidder) {
					t.Fatalf("expected bidders is missing bidder %s", actualBidder.GUID)
				}
				if len(actual) != len(test.expected) {
					t.Fatalf("size diff in filtered bidders")
				}
			}
		})
	}
}
