package multishard

type (
	ChunkIdx  int
	ServerIdx int
	ShardMap  map[ChunkIdx]ServerIdx
)
