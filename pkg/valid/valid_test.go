package valid

import "testing"

func TestOneOf(t *testing.T) {
	type args struct {
		values []interface{}
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "check empty list",
			args: args{
				values: []interface{}{},
			},
			want: false,
		},
		{
			name: "check one",
			args: args{
				values: []interface{}{"test"},
			},
			want: true,
		},
		{
			name: "check one with nil",
			args: args{
				values: []interface{}{"test", nil, nil},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := OneOf(tt.args.values...); got != tt.want {
				t.Errorf("OneOf() = %v, want %v", got, tt.want)
			}
		})
	}
}
