package utils

import (
	"strconv"
	"strings"
	"testing"
)

func TestGenerateOrderNoIncludesMerchantIDAndRandomDigits(t *testing.T) {
	merchantID := uint64(12345)
	orderNo := GenerateOrderNo(merchantID)

	if len(orderNo) != 25 {
		t.Fatalf("expected order number length 25, got %d: %s", len(orderNo), orderNo)
	}

	if !strings.Contains(orderNo, strconv.FormatUint(merchantID, 10)) {
		t.Fatalf("expected order number to contain merchant id %d, got %s", merchantID, orderNo)
	}
}

func TestGenerateOrderNoTrimsVeryLongMerchantID(t *testing.T) {
	merchantID := uint64(1234567890123456)
	orderNo := GenerateOrderNo(merchantID)

	if len(orderNo) > 32 {
		t.Fatalf("expected order number length <= 32, got %d: %s", len(orderNo), orderNo)
	}

	if !strings.Contains(orderNo, "567890123456") {
		t.Fatalf("expected trimmed merchant id suffix in order number, got %s", orderNo)
	}
}
