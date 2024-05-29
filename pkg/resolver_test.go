package resolver

import (
	"errors"
	"testing"
	"time"
)

func justTime(dateTime string) time.Time {
	parsedTime, _ := time.Parse(time.DateTime, dateTime)
	return parsedTime
}

func TestAuctionResolver(t *testing.T) {
	tests := map[string]struct {
		bidders                   []Bidder
		expectedWinnerGUID        string
		expectedWinnerAmountCents int64
		auction                   Auction
		expectedErr               error
	}{
		"happy path": {
			auction: Auction{
				"testAuction",
				100,
			},
			bidders: []Bidder{
				Bidder{
					GUID: "Low",
					BidStrategy: BidStrategy{
						MaxBidCents:       2000,
						BidIncrementCents: 500,
					},
					InitialBid: Bid{
						AmountCents: 1000,
						CreatedAt:   justTime("2006-01-01 01:00:00"),
					},
				},
				Bidder{
					GUID: "High",
					BidStrategy: BidStrategy{
						MaxBidCents:       30000,
						BidIncrementCents: 500,
					},
					InitialBid: Bid{
						AmountCents: 500,
						CreatedAt:   justTime("2007-01-01 01:00:00"),
					},
				},
			},
			expectedWinnerGUID:        "High",
			expectedWinnerAmountCents: 2500,
		},
		"happy path, tie between two of larger auction": {
			auction: Auction{
				"testAuction",
				100,
			},
			bidders: []Bidder{
				Bidder{
					GUID: "EarlyWinner",
					BidStrategy: BidStrategy{
						MaxBidCents:       3000,
						BidIncrementCents: 500,
					},
					InitialBid: Bid{
						AmountCents: 1000,
						CreatedAt:   justTime("2006-01-01 01:00:00"),
					},
				},
				Bidder{
					GUID: "LateWinner",
					BidStrategy: BidStrategy{
						MaxBidCents:       3000,
						BidIncrementCents: 500,
					},
					InitialBid: Bid{
						AmountCents: 500,
						CreatedAt:   justTime("2007-01-01 01:00:00"),
					},
				},
				Bidder{
					GUID: "High",
					BidStrategy: BidStrategy{
						MaxBidCents:       120,
						BidIncrementCents: 1,
					},
					InitialBid: Bid{
						AmountCents: 50,
						CreatedAt:   justTime("2007-01-01 01:00:00"),
					},
				},
				Bidder{
					GUID: "High",
					BidStrategy: BidStrategy{
						MaxBidCents:       100,
						BidIncrementCents: 1,
					},
					InitialBid: Bid{
						AmountCents: 50,
						CreatedAt:   justTime("2007-01-01 01:00:00"),
					},
				},
			},
			expectedWinnerGUID:        "EarlyWinner",
			expectedWinnerAmountCents: 3000,
		},
		"no bidders": {
			auction: Auction{
				"testAuction",
				100,
			},
			bidders:                   []Bidder{},
			expectedWinnerGUID:        "",
			expectedWinnerAmountCents: 0,
			expectedErr:               errors.New("no bidders"),
		},
		"one bidder, bidderMax < auctionMin": {
			auction: Auction{
				"testAuction",
				10000,
			},
			bidders: []Bidder{
				Bidder{
					GUID: "Cheapskate",
					BidStrategy: BidStrategy{
						MaxBidCents:       2000,
						BidIncrementCents: 500,
					},
					InitialBid: Bid{
						AmountCents: 1000,
						CreatedAt:   justTime("2006-01-01 01:00:00"),
					},
				},
			},
			expectedWinnerGUID:        "",
			expectedWinnerAmountCents: 0,
			expectedErr:               errors.New("all bidders too cheap"),
		},
		"one bidder, bidderMax > auctionMin": {
			auction: Auction{
				"testAuction",
				1800,
			},
			bidders: []Bidder{
				Bidder{
					GUID: "Low",
					BidStrategy: BidStrategy{
						MaxBidCents:       2000,
						BidIncrementCents: 20,
					},
					InitialBid: Bid{
						AmountCents: 1000,
						CreatedAt:   justTime("2006-01-01 01:00:00"),
					},
				},
			},
			expectedWinnerGUID:        "Low",
			expectedWinnerAmountCents: 1800,
		},
		"three bidders, tied, winner is earliest bidder": {
			auction: Auction{
				"testAuction",
				2000,
			},
			bidders: []Bidder{
				Bidder{
					GUID: "EarlyBird",
					BidStrategy: BidStrategy{
						MaxBidCents:       2000,
						BidIncrementCents: 20,
					},
					InitialBid: Bid{
						AmountCents: 1000,
						CreatedAt:   justTime("2006-01-01 01:00:00"),
					},
				},
				Bidder{
					GUID: "HungryBird",
					BidStrategy: BidStrategy{
						MaxBidCents:       2000,
						BidIncrementCents: 20,
					},
					InitialBid: Bid{
						AmountCents: 1000,
						CreatedAt:   justTime("2007-01-01 01:00:00"),
					},
				},
				Bidder{
					GUID: "SleepyBird",
					BidStrategy: BidStrategy{
						MaxBidCents:       2000,
						BidIncrementCents: 20,
					},
					InitialBid: Bid{
						AmountCents: 1000,
						CreatedAt:   justTime("2008-01-01 01:00:00"),
					},
				},
			},
			expectedWinnerGUID:        "EarlyBird",
			expectedWinnerAmountCents: 2000,
		},
	}
	for name, tt := range tests {
		test := tt
		t.Run(name, func(t *testing.T) {
			winner, winningAmount, err := ResolveAuction(
				test.auction,
				test.bidders,
			)

			if winner != test.expectedWinnerGUID {
				t.Fatalf("fail: winner expected %s, got %s", test.expectedWinnerGUID, winner)
			}

			if winningAmount != test.expectedWinnerAmountCents {
				t.Fatalf("fail: amount expected %d, got %d", test.expectedWinnerAmountCents, winningAmount)
			}

			if err != nil {
				if test.expectedErr == nil {
					t.Fatalf("fail: unexpected err %s", err.Error())
				}
			}

			if test.expectedErr != nil {
				if err == nil {
					t.Fatalf("fail, didnt catch expected err %s", test.expectedErr.Error())
				}
				if err.Error() != test.expectedErr.Error() {
					t.Fatalf("err mismatch, expecteed %s got %s", test.expectedErr.Error(), err.Error())
				}
			}

		})
	}
}
