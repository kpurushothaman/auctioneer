package resolver

import (
	"errors"
	"slices"
	"time"
)

type Bid struct {
	CreatedAt   time.Time
	AmountCents int64
}

type BidStrategy struct {
	MaxBidCents       int64
	BidIncrementCents int64
}

type Bidder struct {
	GUID        string
	BidStrategy BidStrategy
	InitialBid  Bid
}

// compareBiddersByBidTime is used to determine which bidder placed their initial bid first
func compareBiddersByBidTime(a, b Bidder) int {
	if a.InitialBid.CreatedAt.Before(b.InitialBid.CreatedAt) {
		return -1
	} else if a.InitialBid.CreatedAt.After(b.InitialBid.CreatedAt) {
		return 1
	}
	return 0
}

func sortBiddersByEarliest(bidders []Bidder) {
	slices.SortStableFunc(bidders, func(a, b Bidder) int {
		return compareBiddersByBidTime(a, b)
	})
}

func removeBiddersUnderMin(bidders []Bidder, minAmountCents int64) []Bidder {
	return slices.DeleteFunc(bidders, func(bidder Bidder) bool {
		return bidder.BidStrategy.MaxBidCents < minAmountCents
	})
}

func getInitialBids(bidders []Bidder) (bids map[string]int64, highestBidder string, err error) {
	initialBids := map[string]int64{}
	highestBid := int64(0)
	highestBidder = ""

	if !slices.IsSortedFunc(bidders, compareBiddersByBidTime) {
		return initialBids, "", errors.New("getInitialBids(): passed array not sorted")
	}

	for _, bidder := range bidders {
		initialBids[bidder.GUID] = bidder.InitialBid.AmountCents

		if highestBid < bidder.InitialBid.AmountCents {
			highestBid = bidder.InitialBid.AmountCents
			highestBidder = bidder.GUID
		}
	}

	return initialBids, highestBidder, nil
}

func getEarliestBidderAtMaxValue(bidders []Bidder) (bidValueToEarliestGUID map[int64]string, err error) {
	maxBids := map[int64]string{}

	if !slices.IsSortedFunc(bidders, compareBiddersByBidTime) {
		return maxBids, errors.New("pass sorted arr to maxBids func")
	}

	for _, bidder := range bidders {
		_, ok := maxBids[bidder.BidStrategy.MaxBidCents]
		if !ok {
			maxBids[bidder.BidStrategy.MaxBidCents] = bidder.GUID
		}
	}

	return maxBids, nil
}
