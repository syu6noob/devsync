package server

import (
	"sort"

	"github.com/syu6noob/devsync/internal/core"
)

func BuildPlan(req core.SyncRequest, serverFiles map[string]core.FileMeta) core.SyncPlan {
	keysMap := map[string]struct{}{}
	for p := range req.Files {
		keysMap[p] = struct{}{}
	}
	for p := range req.Base {
		keysMap[p] = struct{}{}
	}
	for p, m := range serverFiles {
		if !m.IsDeleted {
			keysMap[p] = struct{}{}
		}
	}
	keys := make([]string, 0, len(keysMap))
	for p := range keysMap {
		keys = append(keys, p)
	}
	sort.Strings(keys)

	ops := []core.Operation{}
	for _, p := range keys {
		local, hasLocal := req.Files[p]
		base, hasBase := req.Base[p]
		server, hasServer := serverFiles[p]
		serverActive := hasServer && !server.IsDeleted

		switch {
		case hasLocal && serverActive:
			if local.SHA256 == server.SHA256 {
				continue
			}
			if hasBase && server.Version > base.Version && local.SHA256 != base.SHA256 {
				ops = append(ops, core.Operation{Action: "conflict", Path: p, Meta: server, Reason: "local and remote changed"})
			} else {
				ops = append(ops, core.Operation{Action: "upload", Path: p, Meta: local, Reason: "local differs from remote"})
			}

		case hasLocal && !serverActive:
			if hasServer && hasBase && server.Version > base.Version {
				ops = append(ops, core.Operation{Action: "conflict", Path: p, Meta: server, Reason: "remote deleted but local changed"})
			} else {
				ops = append(ops, core.Operation{Action: "upload", Path: p, Meta: local, Reason: "new local file"})
			}

		case !hasLocal && serverActive:
			if hasBase {
				if server.Version > base.Version {
					ops = append(ops, core.Operation{Action: "download", Path: p, Meta: server, Reason: "remote changed"})
				} else {
					ops = append(ops, core.Operation{Action: "delete_remote", Path: p, Meta: base, Reason: "locally deleted"})
				}
			} else {
				ops = append(ops, core.Operation{Action: "download", Path: p, Meta: server, Reason: "new remote file"})
			}
		}
	}
	return core.SyncPlan{Operations: ops}
}
