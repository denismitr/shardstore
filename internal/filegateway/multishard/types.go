package multishard

type (
	ChunkIDX   int
	ServerIDX  int
	MultiShard map[ChunkIDX]ServerIDX
)
