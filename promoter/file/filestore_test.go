/*
Copyright 2021 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package file

import (
	"testing"

	"github.com/stretchr/testify/require"

	api "sigs.k8s.io/promo-tools/v4/api/files"
)

func Test_useStorageClientAuth(t *testing.T) {
	type args struct {
		filestore         *api.Filestore
		useServiceAccount bool
		confirm           bool
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "production",
			args: args{
				filestore: &api.Filestore{
					ServiceAccount: "good@service.account",
				},
				confirm: true,
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "production without service account",
			args: args{
				filestore: &api.Filestore{},
				confirm:   true,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "production source filestore without service account",
			args: args{
				filestore: &api.Filestore{
					Src: true,
				},
				confirm: true,
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "non-production",
			args: args{
				filestore: &api.Filestore{},
				confirm:   false,
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "non-production with service account failure",
			args: args{
				filestore:         &api.Filestore{},
				confirm:           false,
				useServiceAccount: true,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "non-production with service account success",
			args: args{
				filestore: &api.Filestore{
					ServiceAccount: "good@service.account",
				},
				confirm:           false,
				useServiceAccount: true,
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := useStorageClientAuth(
				tt.args.filestore,
				tt.args.useServiceAccount,
				tt.args.confirm,
			)

			if tt.wantErr {
				require.Error(t, err)
			}

			require.Equal(t, tt.want, got)
		},
		)
	}
}
