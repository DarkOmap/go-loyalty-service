package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

type Order struct {
	Number     int64      `json:"number"`
	Status     string     `json:"status"`
	Accrual    *float64   `json:"accrual,omitempty"`
	UploadedAt *time.Time `json:"uploaded_at"`
}

func (o *Order) ScanRow(rows pgx.Rows) error {
	values, err := rows.Values()
	if err != nil {
		return err
	}

	for i := range values {
		switch strings.ToLower(rows.FieldDescriptions()[i].Name) {
		case "number":
			o.Number = values[i].(int64)
		case "status":
			o.Status = values[i].(string)
		case "accrual":
			acc := values[i]

			if acc != nil {
				acc := acc.(float64)
				o.Accrual = &acc
			}
		case "uploadedat":
			ua := values[i].(time.Time)
			o.UploadedAt = &ua
		}
	}

	return nil
}

func (o *Order) UnmarshalJSON(data []byte) (err error) {
	type OrderFromService struct {
		Order   string   `json:"order"`
		Status  string   `json:"status"`
		Accrual *float64 `json:"accrual,omitempty"`
	}

	ofs := &OrderFromService{}

	if err = json.Unmarshal(data, ofs); err != nil || ofs.Order == "" {
		err = json.Unmarshal(data, o)
		return
	}

	o.Number, err = strconv.ParseInt(ofs.Order, 10, 64)
	o.Status = ofs.Status

	if ofs.Status == "REGISTERED" {
		o.Status = "NEW"
	}

	o.Accrual = ofs.Accrual

	return
}

func (o Order) MarshalJSON() ([]byte, error) {
	type OrderAlias Order

	aliasOrder := struct {
		OrderAlias
		Number string `json:"number"`
	}{
		OrderAlias: OrderAlias(o),
		Number:     strconv.Itoa(int(o.Number)),
	}

	return json.Marshal(aliasOrder)
}

type OrderBalance struct {
	Order       int64      `json:"order"`
	Sum         float64    `json:"sum"`
	ProcessedAt *time.Time `json:"processed_at"`
}

func NewOrderBalanceByRequestBody(body io.ReadCloser) (*OrderBalance, error) {
	var buf bytes.Buffer
	_, err := buf.ReadFrom(body)

	if err != nil {
		return nil, fmt.Errorf("read from body: %w", err)
	}

	var ob OrderBalance
	err = ob.writeFieldsByJSON(buf.Bytes())

	if err != nil {
		return nil, fmt.Errorf("write fields for order balance: %w", err)
	}

	return &ob, nil
}

func (ob *OrderBalance) ScanRow(rows pgx.Rows) error {
	values, err := rows.Values()
	if err != nil {
		return err
	}

	for i := range values {
		switch strings.ToLower(rows.FieldDescriptions()[i].Name) {
		case "order":
			ob.Order = values[i].(int64)
		case "sum":
			ob.Sum = values[i].(float64)
		case "processedat":
			pa := values[i].(time.Time)
			ob.ProcessedAt = &pa
		}
	}

	return nil
}

func (ob *OrderBalance) UnmarshalJSON(data []byte) (err error) {
	type OrderBalanceAlias OrderBalance

	aliasOrderBalance := &struct {
		*OrderBalanceAlias
		Order string `json:"order"`
	}{
		OrderBalanceAlias: (*OrderBalanceAlias)(ob),
	}

	if err = json.Unmarshal(data, aliasOrderBalance); err != nil {
		return
	}
	ob.Order, err = strconv.ParseInt(aliasOrderBalance.Order, 10, 64)
	return
}

func (ob OrderBalance) MarshalJSON() ([]byte, error) {
	type OrderBalanceAlias OrderBalance

	aliasOrderBalance := struct {
		OrderBalanceAlias
		Order string `json:"order"`
	}{
		OrderBalanceAlias: OrderBalanceAlias(ob),
		Order:             strconv.Itoa(int(ob.Order)),
	}

	return json.Marshal(aliasOrderBalance)
}

func (ob *OrderBalance) writeFieldsByJSON(j []byte) error {
	err := json.Unmarshal(j, ob)

	if err != nil {
		return fmt.Errorf("unmarshall json %s: %w", string(j), err)
	}

	return nil
}
