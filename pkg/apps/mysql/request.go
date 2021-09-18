package mysql

import (
	"errors"
	"fmt"
	"github.com/jixindatech/pkg/config"
	"github.com/jixindatech/pkg/core/golog"
	"go.uber.org/zap"
	"time"
)

func processRequest(info *MysqlInfo, data []byte) error {
	cmd := data[0]
	data = data[1:]

	switch cmd {
	case COM_QUIT:
		return nil
	case COM_QUERY:
		var sqlData string
		if data[len(data)-1] == ';' {
			sqlData = string(data[:len(data)-1])
		} else {
			sqlData = string(data)
		}
		info.Sql = sqlData

		now := time.Now().Unix()
		msg := config.SqlMsg{
			Src:   info.Src,
			Dst:   info.Dst,
			User:  info.User,
			Time:  now,
			Db:    info.Db,
			Sql:   sqlData,
			Error: PARSE_ERROR,
			Type:  TYPE_UNKNOWN,
			Op:    OP_UNKNOWN,
		}

		var err error
		var res string
		info.Op, res, err = getSqlOp(sqlData)
		if err != nil {
			golog.Error("op", zap.String("err", err.Error()))
		}

		msg.Op = info.Op

		if info.Op != OP_UNKNOWN {
			msg.Error = PARSE_OK
			if info.Op == OP_USE {
				info.Db = res
			}
		}

		ruleId, ruleType, ruleName, ruleAlert := matchRules(info, &ruleConfig)
		if ruleId != 0 {
			msg.Name = ruleName
			msg.Id = ruleId
			msg.Type = ruleType
			msg.Alert = ruleAlert
		}

		err = info.Queue.Put(msg)

	case COM_PING:
	case COM_INIT_DB:
		info.Db = string(data)
	case COM_FIELD_LIST:
	case COM_STMT_PREPARE:
	case COM_STMT_EXECUTE:
	case COM_STMT_CLOSE:
	case COM_STMT_SEND_LONG_DATA:
	case COM_STMT_RESET:
	case COM_SET_OPTION:
	default:
		msg := fmt.Sprintf("command %d not supported now", cmd)
		return errors.New(msg)
	}

	return nil
}
