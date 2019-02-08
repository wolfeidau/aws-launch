package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_New(t *testing.T) {

	got := New(nil)
	require.NotNil(t, got)

	// type args struct {
	// 	cfgs []*aws.Config
	// }
	// tests := []struct {
	// 	name string
	// 	args args
	// 	want *ServiceDispatcher
	// }{
	// 	{
	// 		args: args{
	// 			cfgs: nil,
	// 		},
	// 		want: &ServiceDispatcher{

	// 		},
	// 	},
	// }
	// for _, tt := range tests {
	// 	t.Run(tt.name, func(t *testing.T) {
	// 		if got := New(tt.args.cfgs...); !reflect.DeepEqual(got, tt.want) {
	// 			t.Errorf("New() = %v, want %v", got, tt.want)
	// 		}
	// 	})
	// }
}

func TestServiceDispatcher_DefineAndLaunch(t *testing.T) {

}
