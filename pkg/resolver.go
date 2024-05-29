package resolver

import (
	"errors"
	"log/slog"
	"slices"
)

type Auction struct {
	id            string
	minPriceCents int64
}

// ResolveAuction takes in an auction and a set of bidders. It returns the GUID of the winner, their winning bid, and an error if one was encountered.
func ResolveAuction(auction Auction, bidders []Bidder) (string, int64, error) {
	if len(bidders) == 0 {
		return "", 0, errors.New("no bidders")
	}

	sortBiddersByEarliest(bidders)

	// do an initial cleaning pass against the bidder pool -- this cuts down on the number of riffraff we have to
	// expend compute resources on
	bidders = removeBiddersUnderMin(bidders, auction.minPriceCents)
	if len(bidders) == 0 {
		return "", 0, errors.New("all bidders too cheap")
	}

	bids, highestBidder, err := getInitialBids(bidders)
	if err != nil {
		slog.Error(err.Error())
	}

	// store a map of highestBid:bidderWithEarliestInitialBidForThatValue to handle ties
	maxBidWinners, err := getEarliestBidderAtMaxValue(bidders)
	if err != nil {
		slog.Error(err.Error())
	}

	// at this point, we have a pool of bidders willing to pay the min price. Let's run the auction.
	eligibleBidders := true
	for eligibleBidders == true {
		// terminal condition for a tied auction -- one bidder hits the max, and then all bidders are removed either because they can't hit the max
		// or their next bid would be over their max bid.
		if len(bidders) == 0 {
			eligibleBidders = false
			continue
		}

		for _, bidder := range bidders {
			// terminal condition for non-tied auction
			// while the only eligible bidder is under the min price, run winner's strategy until we exceed the min price
			if len(bidders) == 1 {
				if bids[highestBidder] < auction.minPriceCents {
					bids[highestBidder] += bidders[0].BidStrategy.BidIncrementCents
				} else {
					eligibleBidders = false
				}
				continue
			}

			// don't increment winner's bid if they're winning
			if highestBidder == bidder.GUID {
				continue
			}

			highestBid, ok := bids[highestBidder]
			if !ok {
				// actually not sure how we could possibly get this condition, but panics aren't any fun, so cover it
				return "", 0, errors.New("bidsMap initialized incorrectly, missing highest bidder bid")
			}

			nextWinningBidForBidder := highestBid + bidder.BidStrategy.BidIncrementCents

			if nextWinningBidForBidder > highestBid && nextWinningBidForBidder <= bidder.BidStrategy.MaxBidCents {
				bids[bidder.GUID] = nextWinningBidForBidder
				highestBidder = bidder.GUID
			}
		}

		// remove bidders whose max would be under or equal to the winning bid
		// or whose next bid would be over their max bid
		bidders = slices.DeleteFunc(bidders, func(bidder Bidder) bool {
			return bidder.BidStrategy.MaxBidCents <= bids[highestBidder] || bidder.BidStrategy.BidIncrementCents+bids[bidder.GUID] > bidder.BidStrategy.MaxBidCents
		})
	}

	// if we have a tie during the auction, return the earliest bidder who was willing to bid that amount
	// this feels....wildly inelegant but ran out of time
	earliestMaxWinner, ok := maxBidWinners[bids[highestBidder]]
	if ok {
		return earliestMaxWinner, bids[highestBidder], nil
	}

	return highestBidder, bids[highestBidder], nil
}
