package dbase

import (
	"testing"
)

func Test_AllocationBitMap_IsAllocated(t *testing.T) {

	var (
		bytesA = []byte{255, 254, 253, 251, 247, 239, 223, 191, 127, 1, 2, 4, 8, 16, 32, 64, 128}
		// {11111111, 11111110, 11111101, 11111011, 11110111, 11101111, 11011111, 10111111, 01111111, 00000001, 00000010, 00000100, 00001000, 00010000, 00100000, 01000000, 10000000}
		//  0:8       8:16      16:24     24:32     32:40     40:48     48:56     56:64     64:72     72:80     80:88     88:96     96:104    104:112   112:120   120:128   128:136
		bitmap = NewAllocationBitMap(bytesA)
	)

	type args struct {
		offset  int
		runsize int
	}
	tests := []struct {
		name    string
		bitmap  AllocationBitMap
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
		{name: "0:8",
			bitmap:  bitmap,
			args:    args{0, 8},
			want:    true,
			wantErr: false,
		},
		{name: "0:9",
			bitmap:  bitmap,
			args:    args{0, 9},
			want:    true,
			wantErr: false,
		},
		{name: "8:1",
			bitmap:  bitmap,
			args:    args{8, 1},
			want:    false,
			wantErr: false,
		},
		{name: "9:7",
			bitmap:  bitmap,
			args:    args{9, 7},
			want:    true,
			wantErr: false,
		},
		{name: "64:7",
			bitmap:  bitmap,
			args:    args{64, 7},
			want:    true,
			wantErr: false,
		},
		{name: "91:8",
			bitmap:  bitmap,
			args:    args{91, 8},
			want:    false,
			wantErr: false,
		},
		{name: "128:7",
			bitmap:  bitmap,
			args:    args{128, 7},
			want:    false,
			wantErr: false,
		},
		{name: "128:8",
			bitmap:  bitmap,
			args:    args{128, 8},
			want:    true,
			wantErr: false,
		},
		{name: "128:16",
			bitmap:  bitmap,
			args:    args{128, 16},
			want:    false,
			wantErr: true,
		},
		{name: "10000:10000",
			bitmap:  bitmap,
			args:    args{10000, 10000},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.bitmap.IsAllocated(tt.args.offset, tt.args.runsize)
			if (err != nil) != tt.wantErr {
				t.Errorf("AllocationBitMap.IsAllocated() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("AllocationBitMap.IsAllocated() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_AllocationBitMap_AllocateExplicit(t *testing.T) {

	var (
		bytesA = []byte{255, 254, 253, 251, 247, 239, 223, 191, 127, 1, 2, 4, 8, 16, 32, 64, 128}
		// {11111111, 11111110, 11111101, 11111011, 11110111, 11101111, 11011111, 10111111, 01111111, 00000001, 00000010, 00000100, 00001000, 00010000, 00100000, 01000000, 10000000}
		//  0:8       8:16      16:24     24:32     32:40     40:48     48:56     56:64     64:72     72:80     80:88     88:96     96:104    104:112   112:120   120:128   128:136
		bitmap = NewAllocationBitMap(bytesA)
	)

	type args struct {
		offset  int
		runsize int
	}
	tests := []struct {
		name          string
		bitmap        AllocationBitMap
		args          args
		wantBefore    bool
		wantErrBefore bool
		wantErr       bool
		wantAfter     bool
		wantErrAfter  bool
	}{
		// TODO: Add test cases.
		{name: "0:8",
			bitmap:        bitmap,
			args:          args{0, 8},
			wantBefore:    true,
			wantErrBefore: false,
			wantErr:       false,
			wantAfter:     true,
			wantErrAfter:  false,
		},
		{name: "0:9",
			bitmap:        bitmap,
			args:          args{0, 9},
			wantBefore:    true,
			wantErrBefore: false,
			wantErr:       false,
			wantAfter:     true,
			wantErrAfter:  false,
		},
		{name: "8:1",
			bitmap:        bitmap,
			args:          args{8, 1}, // NB, already mutated from 0:9 above
			wantBefore:    true,
			wantErrBefore: false,
			wantErr:       false,
			wantAfter:     true,
			wantErrAfter:  false,
		},
		{name: "9:7",
			bitmap:        bitmap,
			args:          args{9, 7},
			wantBefore:    true,
			wantErrBefore: false,
			wantErr:       false,
			wantAfter:     true,
			wantErrAfter:  false,
		},
		{name: "64:7",
			bitmap:        bitmap,
			args:          args{64, 7},
			wantBefore:    true,
			wantErrBefore: false,
			wantErr:       false,
			wantAfter:     true,
			wantErrAfter:  false,
		},
		{name: "91:8",
			bitmap:        bitmap,
			args:          args{91, 8},
			wantBefore:    false,
			wantErrBefore: false,
			wantErr:       false,
			wantAfter:     true,
			wantErrAfter:  false,
		},
		{name: "128:7",
			bitmap:        bitmap,
			args:          args{128, 7},
			wantBefore:    false,
			wantErrBefore: false,
			wantErr:       false,
			wantAfter:     true,
			wantErrAfter:  false,
		},
		{name: "128:8",
			bitmap:        bitmap,
			args:          args{128, 8},
			wantBefore:    true,
			wantErrBefore: false,
			wantErr:       false,
			wantAfter:     true,
			wantErrAfter:  false,
		},
		{name: "128:16",
			bitmap:        bitmap,
			args:          args{128, 16},
			wantBefore:    false,
			wantErrBefore: true,
			wantErr:       true,
			wantAfter:     false,
			wantErrAfter:  true,
		},
		{name: "10000:10000",
			bitmap:        bitmap,
			args:          args{10000, 10000},
			wantBefore:    false,
			wantErrBefore: true,
			wantErr:       true,
			wantAfter:     false,
			wantErrAfter:  true,
		},
		{name: "0:136",
			bitmap:        bitmap,
			args:          args{0, 136},
			wantBefore:    true,
			wantErrBefore: false,
			wantErr:       false,
			wantAfter:     true,
			wantErrAfter:  false,
		},
		{name: "0:64000",
			bitmap:        NewAllocationBitMap(make([]byte, 8*1000)),
			args:          args{0, 64000},
			wantBefore:    false,
			wantErrBefore: false,
			wantErr:       false,
			wantAfter:     true,
			wantErrAfter:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.bitmap.IsAllocated(tt.args.offset, tt.args.runsize)
			if (err != nil) != tt.wantErrBefore {
				t.Errorf("Before AllocateExplicit: AllocationBitMap.IsAllocated() error = %v, wantErrBefore %v", err, tt.wantErrBefore)
				return
			}
			if got != tt.wantBefore {
				t.Errorf("Before AllocateExplicit: AllocationBitMap.IsAllocated() = %v, want %v", got, tt.wantBefore)
			}
			if err = tt.bitmap.AllocateExplicit(tt.args.offset, tt.args.runsize); (err != nil) != tt.wantErr {
				t.Errorf("AllocationBitMap.AllocateExplicit() error = %v, wantErr %v", err, tt.wantErr)
			}
			got, err = tt.bitmap.IsAllocated(tt.args.offset, tt.args.runsize)
			if (err != nil) != tt.wantErrAfter {
				t.Errorf("After AllocateExplicit: AllocationBitMap.IsAllocated() error = %v, wantErr %v", err, tt.wantErrAfter)
				return
			}
			if got != tt.wantAfter {
				t.Errorf("After AllocateExplicit: AllocationBitMap.IsAllocated() = %v, want %v", got, tt.wantAfter)
			}
		})
		//fmt.Printf("bytes: %v %v\n", len(tt.bitmap.GetBytes()), tt.bitmap.GetBytes())
	}
	//fmt.Printf("len %v, bytes: %v\n"bytesA)
}

func Test_AllocationBitMap_Deallocate(t *testing.T) {

	var (
		bytesA = []byte{255, 254, 253, 251, 247, 239, 223, 191, 127, 1, 2, 4, 8, 16, 32, 64, 128}
		// {11111111, 11111110, 11111101, 11111011, 11110111, 11101111, 11011111, 10111111, 01111111, 00000001, 00000010, 00000100, 00001000, 00010000, 00100000, 01000000, 10000000}
		//  0:8       8:16      16:24     24:32     32:40     40:48     48:56     56:64     64:72     72:80     80:88     88:96     96:104    104:112   112:120   120:128   128:136
		bitmap = NewAllocationBitMap(bytesA)
	)

	type args struct {
		offset  int
		runsize int
	}
	tests := []struct {
		name          string
		bitmap        AllocationBitMap
		args          args
		wantBefore    bool
		wantErrBefore bool
		wantErr       bool
		wantAfter     bool
		wantErrAfter  bool
	}{
		{name: "0:8",
			bitmap:        bitmap,
			args:          args{0, 8},
			wantBefore:    true,
			wantErrBefore: false,
			wantErr:       false,
			wantAfter:     false,
			wantErrAfter:  false,
		},
		{name: "0:9",
			bitmap:        bitmap,
			args:          args{0, 9},
			wantBefore:    false,
			wantErrBefore: false,
			wantErr:       false,
			wantAfter:     false,
			wantErrAfter:  false,
		},
		{name: "8:1",
			bitmap:        bitmap,
			args:          args{8, 1}, // NB, already mutated from 0:9 above
			wantBefore:    false,
			wantErrBefore: false,
			wantErr:       false,
			wantAfter:     false,
			wantErrAfter:  false,
		},
		{name: "9:7",
			bitmap:        bitmap,
			args:          args{9, 7},
			wantBefore:    true,
			wantErrBefore: false,
			wantErr:       false,
			wantAfter:     false,
			wantErrAfter:  false,
		},
		{name: "64:7",
			bitmap:        bitmap,
			args:          args{64, 7},
			wantBefore:    true,
			wantErrBefore: false,
			wantErr:       false,
			wantAfter:     false,
			wantErrAfter:  false,
		},
		{name: "91:8",
			bitmap:        bitmap,
			args:          args{91, 8},
			wantBefore:    false,
			wantErrBefore: false,
			wantErr:       false,
			wantAfter:     false,
			wantErrAfter:  false,
		},
		{name: "128:7",
			bitmap:        bitmap,
			args:          args{128, 7},
			wantBefore:    false,
			wantErrBefore: false,
			wantErr:       false,
			wantAfter:     false,
			wantErrAfter:  false,
		},
		{name: "128:8",
			bitmap:        bitmap,
			args:          args{128, 8},
			wantBefore:    true,
			wantErrBefore: false,
			wantErr:       false,
			wantAfter:     false,
			wantErrAfter:  false,
		},
		{name: "128:16",
			bitmap:        bitmap,
			args:          args{128, 16},
			wantBefore:    false,
			wantErrBefore: true,
			wantErr:       true,
			wantAfter:     false,
			wantErrAfter:  true,
		},
		{name: "10000:10000",
			bitmap:        bitmap,
			args:          args{10000, 10000},
			wantBefore:    false,
			wantErrBefore: true,
			wantErr:       true,
			wantAfter:     false,
			wantErrAfter:  true,
		},
		{name: "0:136",
			bitmap:        bitmap,
			args:          args{0, 136},
			wantBefore:    true,
			wantErrBefore: false,
			wantErr:       false,
			wantAfter:     false,
			wantErrAfter:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.bitmap.IsAllocated(tt.args.offset, tt.args.runsize)
			if (err != nil) != tt.wantErrBefore {
				t.Errorf("Before Deallocate: AllocationBitMap.IsAllocated() error = %v, wantErrBefore %v", err, tt.wantErrBefore)
				return
			}
			if got != tt.wantBefore {
				t.Errorf("Before Deallocate: AllocationBitMap.IsAllocated() = %v, want %v", got, tt.wantBefore)
			}
			if err = tt.bitmap.Deallocate(tt.args.offset, tt.args.runsize); (err != nil) != tt.wantErr {
				t.Errorf("AllocationBitMap.Deallocate() error = %v, wantErr %v", err, tt.wantErr)
			}
			got, err = tt.bitmap.IsAllocated(tt.args.offset, tt.args.runsize)
			if (err != nil) != tt.wantErrAfter {
				t.Errorf("After Deallocate: AllocationBitMap.IsAllocated() error = %v, wantErr %v", err, tt.wantErrAfter)
				return
			}
			if got != tt.wantAfter {
				t.Errorf("After Deallocate: AllocationBitMap.IsAllocated() = %v, want %v", got, tt.wantAfter)
			}
		})
	}
	//fmt.Printf("bytes: %v\n", bytesA)
}
