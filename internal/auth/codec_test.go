package auth

import (
	"reflect"
	"testing"

	"github.com/gofrs/uuid"
)

func BenchmarkEncodeUID(b *testing.B) {
	uid := uuid.Must(uuid.NewV4())

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = EncodeUID(uid)
	}
}

func BenchmarkDecodeUID(b *testing.B) {
	uid, _ := EncodeUID(uuid.Must(uuid.NewV4()))

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = DecodeUID(uid)
	}
}

func TestDecodeUID(t *testing.T) {
	uuidForTest, _ := uuid.NewV4()
	uuidEncoded, _ := EncodeUID(uuidForTest)
	type args struct {
		ciphertext []byte
	}
	tests := []struct {
		name    string
		args    args
		want    *uuid.UUID
		wantErr bool
	}{
		{"regular decode", args{ciphertext: uuidEncoded}, &uuidForTest, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeUID(tt.args.ciphertext)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeUID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DecodeUID() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecodeUIDFromHex(t *testing.T) {
	uuidForTest, _ := uuid.NewV4()
	hexUUID, _ := EncodeUIDToHex(uuidForTest)
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    *uuid.UUID
		wantErr bool
	}{
		{"regular decode from hex", args{s: hexUUID}, &uuidForTest, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeUIDFromHex(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeUIDFromHex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DecodeUIDFromHex() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEncodeUID(t *testing.T) {
	type args struct {
		uid uuid.UUID
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EncodeUID(tt.args.uid)
			if (err != nil) != tt.wantErr {
				t.Errorf("EncodeUID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EncodeUID() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEncodeUIDToHex(t *testing.T) {
	type args struct {
		uid uuid.UUID
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EncodeUIDToHex(tt.args.uid)
			if (err != nil) != tt.wantErr {
				t.Errorf("EncodeUIDToHex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("EncodeUIDToHex() got = %v, want %v", got, tt.want)
			}
		})
	}
}
