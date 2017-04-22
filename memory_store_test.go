package dbase

import (
	"reflect"
	"testing"
)

func TestNewMemoryStore(t *testing.T) {
	tests := []struct {
		name    string
		want    int64
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name:    "New",
			want:    0,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewMemoryStore()
			if (err != nil) != tt.wantErr {
				t.Errorf("NewMemoryStore() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.Count() != tt.want {
				t.Errorf("NewMemoryStore() = %v, want %v", got.Count(), tt.want)
			}
		})
	}
}

func Test_memoryStore_Get(t *testing.T) {
	type args struct {
		id   PageID
		page Page
	}
	tests := []struct {
		name    string
		store   *memoryStore
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.store.Get(tt.args.id, tt.args.page); (err != nil) != tt.wantErr {
				t.Errorf("memoryStore.Get() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_memoryStore_Set(t *testing.T) {
	type args struct {
		id   PageID
		page Page
	}
	tests := []struct {
		name    string
		store   *memoryStore
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.store.Set(tt.args.id, tt.args.page); (err != nil) != tt.wantErr {
				t.Errorf("memoryStore.Set() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_memoryStore_Append(t *testing.T) {
	type args struct {
		page Page
	}
	tests := []struct {
		name    string
		store   *memoryStore
		args    args
		want    PageID
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.store.Append(tt.args.page)
			if (err != nil) != tt.wantErr {
				t.Errorf("memoryStore.Append() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("memoryStore.Append() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_memoryStore_Wipe(t *testing.T) {
	type args struct {
		id PageID
	}
	tests := []struct {
		name    string
		store   *memoryStore
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.store.Wipe(tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("memoryStore.Wipe() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_memoryStore_Count(t *testing.T) {
	tests := []struct {
		name  string
		store *memoryStore
		want  int64
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.store.Count(); got != tt.want {
				t.Errorf("memoryStore.Count() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_memoryStore_Close(t *testing.T) {
	tests := []struct {
		name    string
		store   *memoryStore
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.store.Close(); (err != nil) != tt.wantErr {
				t.Errorf("memoryStore.Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

