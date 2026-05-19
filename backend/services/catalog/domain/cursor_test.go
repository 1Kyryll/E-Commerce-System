package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCursor_IsZero(t *testing.T) {
	var c Cursor
	if !c.IsZero() {
		t.Error("zero-value Cursor.IsZero() = false, want true")
	}
	c.ID = uuid.New()
	if c.IsZero() {
		t.Error("Cursor with non-nil ID is considered zero")
	}
}

func TestCursor_EncodeDecode_RoundTrip(t *testing.T) {
	want := Cursor{
		CreatedAt: time.Date(2026, 4, 17, 10, 30, 0, 0, time.UTC),
		ID:        uuid.MustParse("11111111-2222-3333-4444-555555555555"),
	}
	enc := want.Encode()
	if enc == "" {
		t.Fatal("Encode returned empty string for non-zero cursor")
	}
	got, err := DecodeCursor(enc)
	if err != nil {
		t.Fatalf("DecodeCursor: %v", err)
	}
	if !got.CreatedAt.Equal(want.CreatedAt) {
		t.Errorf("CreatedAt = %v, want %v", got.CreatedAt, want.CreatedAt)
	}
	if got.ID != want.ID {
		t.Errorf("ID = %v, want %v", got.ID, want.ID)
	}
}

func TestCursor_EncodeZero_ReturnsEmpty(t *testing.T) {
	var c Cursor
	if got := c.Encode(); got != "" {
		t.Errorf("zero Cursor.Encode() = %q, want \"\"", got)
	}
}

func TestDecodeCursor_EmptyString_ReturnsZero(t *testing.T) {
	c, err := DecodeCursor("")
	if err != nil {
		t.Fatalf("DecodeCursor(\"\"): %v", err)
	}
	if !c.IsZero() {
		t.Errorf("decoded empty = %+v, want zero", c)
	}
}

func TestDecodeCursor_Garbage_Errors(t *testing.T) {
	_, err := DecodeCursor("not-base64!!!@#$")
	if err == nil {
		t.Error("expected error for garbage input")
	}
}

func TestDecodeCursor_BadJSON_Errors(t *testing.T) {
	bad := "e25vdCBqc29ufQ=="
	_, err := DecodeCursor(bad)
	if err == nil {
		t.Error("expected error for non-JSON payload")
	}
}
