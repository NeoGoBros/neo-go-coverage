package experiment;





// Prefixes used for contract data storage.
const (
	// prefixNumber contains map from the key to stored number.
	prefixNumber byte = 0x00
)
const (
	keySize = 5
)

// IMPORTANT: GetNumber must begin at line 17
func GetNumber(key []byte) int {
	if len(key) != keySize {
		panic("Invalid key size")
	}
	ctx := storage.GetContext()
	num := storage.Get(ctx, append([]byte{prefixNumber}, key...))
	if num == nil {
		panic("Cannot get number")
	}
	return num.(int)
}
// IMPORTANT: PutNumber must begin at line 29
func PutNumber(key []byte, num int) {
	if len(key) != keySize {
		panic("Invalid key size")
	}
	ctx := storage.GetContext()
	storage.Put(ctx, append([]byte{prefixNumber}, key...), num)
}
