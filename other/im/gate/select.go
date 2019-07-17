package main

import (
	"context"
	"cy/other/im/codec"
	"cy/other/im/inner"
	"hash/fnv"
	"sort"
	"strings"
)

type selectByID struct {
	servers []string
}

func (s *selectByID) Select(ctx context.Context, servicePath, serviceMethod string, args interface{}) string {
	if len(s.servers) == 0 {
		return ""
	}

	req, ok := args.(*codec.MsgPayload)
	if !ok {
		return ""
	}

	sessID := inner.SessionID(req.FromUID, req.ToUID, req.IsBroadCast(), req.IsMultiCast())
	idx := jumpConsistentHash(hashString(sessID), int32(len(s.servers)))
	return s.servers[idx]
}

func (s *selectByID) UpdateServer(servers map[string]string) {
	var ss = make([]string, 0, len(servers))
	for k := range servers {
		ss = append(ss, k)
	}

	sort.Slice(ss, func(i, j int) bool {
		return strings.Compare(ss[i], ss[j]) <= 0
	})
	s.servers = ss
}

func hashString(str string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(str))
	return h.Sum64()
}

func jumpConsistentHash(key uint64, buckets int32) int32 {
	if buckets <= 0 {
		buckets = 1
	}

	var b, j int64

	for j < int64(buckets) {
		b = j
		key = key*2862933555777941757 + 1
		j = int64(float64(b+1) * (float64(int64(1)<<31) / float64((key>>33)+1)))
	}

	return int32(b)
}
