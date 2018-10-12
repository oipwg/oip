package oip

import (
	"testing"

	"github.com/bitspill/flod/flojson"
	"github.com/bitspill/oip/datastore"
)

// const floData = "CmUIARIiRlRmcjNWVjFhZEdIQ2lwaEtqZXZhbWd1U2JqckdOZnRDZhm1MPbjkHusPiG5pBzSl9iwPikAAABgiZzfQTH5OD5ZjLVEQjnlM4+yNKbWPkHrsz1ZtZShP0mx0XShUshKQBACGAEiIkZUZnIzVlYxYWRHSENpcGhLamV2YW1ndVNianJHTmZ0Q2YqQR8DnyN4mJRQ9v5P6GKn+ecRIY08dOHeVuhl0kLq7LX5MDxC7r/zf6WxlrJvpkpDq0Iir4ahoR3azjV1jd+DpRtA"
const floData = "CmUIARIiRlRmcjNWVjFhZEdIQ2lwaEtqZXZhbWd1U2JqckdOZnRDZhnXzU3IXmirPiEMi9WSISyzPikAAACAKrzLQTHutJ2vEWlKQjnzhgM69mDgPkG9x7o0VWKoP0nPZtXnardTQBACGAEiIkZUZnIzVlYxYWRHSENpcGhLamV2YW1ndVNianJHTmZ0Q2YqQR/yZ26g/uWHXuIxJZILWXQA/ZSskrcqzccmBg4suALQ7jdj7O1dkPwm3uZ6IXB3mtclubmynAYYX1XfhdeUpUEt"

func TestOnP64(t *testing.T) {
	onP64(floData, &datastore.TransactionData{Transaction: &flojson.TxRawResult{}})
	// ToDo: validate results, only useful in debugger for now
}
