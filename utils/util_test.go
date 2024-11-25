package utils

import (
	"reflect"
	"testing"
)

func TestGetSorted(t *testing.T) {
	tests := []struct {
		name string
		set  *OrderedSet
		desc bool
		want []int64
	}{
		{
			name: "测试升序排序",
			set: &OrderedSet{
				items:   map[int64]struct{}{3: {}, 1: {}, 2: {}},
				isDirty: true,
			},
			desc: false,
			want: []int64{1, 2, 3},
		},
		{
			name: "测试降序排序",
			set: &OrderedSet{
				items:   map[int64]struct{}{3: {}, 1: {}, 2: {}},
				isDirty: true,
			},
			desc: true,
			want: []int64{3, 2, 1},
		},
		{
			name: "测试空集合",
			set: &OrderedSet{
				items:   map[int64]struct{}{},
				isDirty: true,
			},
			desc: false,
			want: []int64{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.set.GetSorted(tt.desc); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSorted() = %v, want %v", got, tt.want)
			}
		})
	}
}
