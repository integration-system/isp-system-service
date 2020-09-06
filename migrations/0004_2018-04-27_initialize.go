package migrations

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"isp-system-service/conf"
	"isp-system-service/controller"
	"isp-system-service/domain"
	path2 "path"

	"github.com/integration-system/isp-lib/v2/config"
	"github.com/integration-system/isp-lib/v2/database"
	"github.com/pressly/goose"
)

const (
	insertDomain  = "INSERT INTO %s.domain (id, name, description, system_id) VALUES (%d, '%s', '%s', %d) RETURNING id;"
	insertService = "INSERT INTO %s.service (id, name, description, domain_id) VALUES (%d, '%s', '%s', %d) RETURNING id;"
	insertApp     = "INSERT INTO %s.application (id, name, description, service_id, type) VALUES (%d, '%s', '%s', %d, '%s') RETURNING id;"
	insertToken   = "INSERT INTO %s.token (token, app_id, expire_time) VALUES ('%s', %d, %d);" //nolint

	setSeqQuery = "SELECT setval('%s.%s_id_seq' :: regclass, %d);"

	initFile = "init.json"
)

func init() {
	goose.AddMigration(Up, Down)
}

func Up(tx *sql.Tx) error {
	path := path2.Join(database.ResolveMigrationsDirectory(), initFile)
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	var list []domain.DomainWithServices
	if err := json.Unmarshal(bytes, &list); err != nil {
		return err
	}

	schema := config.GetRemote().(*conf.RemoteConfig).Database.Schema
	lastDomainId := int64(0)
	lastServiceId := int64(0)
	lastAppId := int64(0)
	for _, domainInfo := range list {
		domainId := int64(0)
		if domainId, err = insertNode(tx, insertDomain, true, schema, domainInfo.Id, domainInfo.Name, domainInfo.Description, 1); err != nil {
			return err
		}
		if domainId > lastDomainId {
			lastDomainId = domainId
		}

		for _, service := range domainInfo.Services {
			serviceId := int64(0)
			if serviceId, err = insertNode(tx, insertService, true, schema, service.Id, service.Name, service.Description, domainId); err != nil {
				return err
			}
			if serviceId > lastServiceId {
				lastServiceId = serviceId
			}

			for _, app := range service.Apps {
				appId := int64(0)
				if appId, err = insertNode(tx, insertApp, true, schema, app.Id, app.Name, app.Description, serviceId, app.Type); err != nil {
					return err
				}
				if appId > lastAppId {
					lastAppId = appId
				}

				for _, t := range app.Tokens {
					if _, err := insertNode(tx, insertToken, false, schema, t.Token, appId, t.ExpireTime); err != nil {
						return err
					} else {
						idMap := map[string]interface{}{
							controller.SystemIdentityFieldInDb:      1,
							controller.DomainIdentityFieldInDb:      domainId,
							controller.ServiceIdentityFieldInDb:     serviceId,
							controller.ApplicationIdentityFieldInDb: appId,
						}
						if err := controller.Token.SetIdentityMapForTokenV2(t.Token, t.ExpireTime, idMap); err != nil {
							return err
						}
					}
				}
			}
		}
	}

	if err := setSeq(tx, schema, "domain", lastDomainId); err != nil {
		return err
	} else if err := setSeq(tx, schema, "service", lastServiceId); err != nil {
		return err
	} else if err := setSeq(tx, schema, "application", lastAppId); err != nil {
		return err
	}

	return nil
}

func Down(tx *sql.Tx) error {
	return nil
}

func insertNode(tx *sql.Tx, query string, scan bool, args ...interface{}) (int64, error) {
	var id int64
	query = fmt.Sprintf(query, args...)
	row := tx.QueryRow(query)
	if scan {
		if err := row.Scan(&id); err != nil {
			return 0, err
		}
	}

	return id, nil
}

func setSeq(tx *sql.Tx, schema, table string, value int64) error {
	q := fmt.Sprintf(setSeqQuery, schema, table, value)
	_, err := tx.Exec(q)

	return err
}
