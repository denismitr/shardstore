package shardmanager

import (
	"github.com/denismitr/shardstore/internal/common/logger"
	"github.com/denismitr/shardstore/internal/filegateway/config"
	"github.com/denismitr/shardstore/internal/filegateway/multishard"
	"reflect"
	"testing"
)

func TestShardManager_GetMultiShard(t *testing.T) {
	type fields struct {
		cfg *config.Config
		lg  logger.Logger
	}

	type args struct {
		key multishard.Key
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    multishard.MultiShard
		wantErr bool
	}{
		{
			name: "3 servers and 3 chunks",
			fields: fields{
				lg: logger.NewStdoutLogger(logger.Dev, "test"),
				cfg: &config.Config{
					StorageServers: make([]string, 3),
					NumberOfChunks: 3,
				},
			},
			wantErr: false,
			want:    multishard.MultiShard{0: 1, 1: 2, 2: 0},
			args: args{
				key: "1_png",
			},
		},
		{
			name: "8 servers and 3 chunks",
			fields: fields{
				lg: logger.NewStdoutLogger(logger.Dev, "test"),
				cfg: &config.Config{
					StorageServers: make([]string, 8),
					NumberOfChunks: 3,
				},
			},
			wantErr: false,
			want:    multishard.MultiShard{0: 7, 1: 0, 2: 1},
			args: args{
				key: "2_png",
			},
		},
		{
			name: "10 servers and 3 chunks different key",
			fields: fields{
				lg: logger.NewStdoutLogger(logger.Dev, "test"),
				cfg: &config.Config{
					StorageServers: make([]string, 10),
					NumberOfChunks: 3,
				},
			},
			wantErr: false,
			want:    multishard.MultiShard{0: 2, 1: 3, 2: 4},
			args: args{
				key: "my_text_txt",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm, err := NewShardManager(tt.fields.cfg, tt.fields.lg)
			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error: %s", err.Error())
			}

			got, err := sm.GetMultiShard(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetMultiShard() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetMultiShard() got = %v, want %v", got, tt.want)
			}
		})
	}
}
