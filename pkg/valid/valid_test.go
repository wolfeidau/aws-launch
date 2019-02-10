package valid

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/stretchr/testify/require"
)

func TestReflectOneOf(t *testing.T) {
	type args struct {
		v     interface{}
		names []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "check one with nil",
			args: args{
				v: struct {
					SS *string
					TT *string
					Bb *string
				}{SS: aws.String("test")}, names: []string{"SS", "TT"},
			},
			want: true,
		},
		{
			name: "check two returns false",
			args: args{
				v: struct {
					SS *string
					TT *string
				}{
					SS: aws.String("test"),
					TT: aws.String("test"),
				},
				names: []string{"SS", "TT"},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ReflectOneOf(tt.args.v, tt.args.names...); got != tt.want {
				t.Errorf("ReflectOneOf() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReflectOneOf_Panic(t *testing.T) {
	require.Panics(t, func() { ReflectOneOf("test", "test") })
}

func TestReflectCountOfNotZero_Panic(t *testing.T) {
	require.Panics(t, func() { ReflectCountOfNotZero("test", "test") })
}
