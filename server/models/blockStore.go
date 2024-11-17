package models

type BlockStore struct {
	Id          int    `json:"id" gorm:"column:id"`
	BlockHeight uint64 `json:"block_height" gorm:"column:block_height"`
}
