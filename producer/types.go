package producer

//Producer generic interface for all prouducers
type Producer interface {
	//Puts a series of records to the producer
	PutRecords([][]byte) [][]byte
}

//PutRecord defaults to calling PutRecords function
func PutRecord(p Producer, record []byte) []byte {
	res := p.PutRecords([][]byte{record})
	if len(res) > 0 {
		return res[0]
	}
	return nil
}
