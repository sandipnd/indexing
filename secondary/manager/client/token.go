// Copyright (c) 2014 Couchbase, Inc.
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
// except in compliance with the License. You may obtain a copy of the License at
//   http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software distributed under the
// License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
// either express or implied. See the License for the specific language governing permissions
// and limitations under the License.

package client

import (
	"encoding/json"
	"errors"
	"fmt"
	c "github.com/couchbase/indexing/secondary/common"
	"github.com/couchbase/indexing/secondary/logging"
)

/////////////////////////////////////////////////////////////////////////
// Const
////////////////////////////////////////////////////////////////////////

const DeleteDDLCommandTokenTag = "commandToken/delete/"
const DDLMetakvDir = c.IndexingMetaDir + "ddl/"
const DeleteDDLCommandTokenPath = DDLMetakvDir + DeleteDDLCommandTokenTag

const IndexerVersionTokenTag = "versionToken"
const InfoMetakvDir = c.IndexingMetaDir + "info/"
const IndexerVersionTokenPath = InfoMetakvDir + IndexerVersionTokenTag

//////////////////////////////////////////////////////////////
// Concrete Type
//////////////////////////////////////////////////////////////

type DeleteCommandToken struct {
	Name   string
	Bucket string
	DefnId c.IndexDefnId
}

type IndexerVersionToken struct {
	Version uint64
}

//////////////////////////////////////////////////////////////
// Delete Token Management
//////////////////////////////////////////////////////////////

//
// Generate a token to metakv for recovery purpose
//
func PostDeleteCommandToken(defnId c.IndexDefnId) error {

	commandToken := &DeleteCommandToken{
		DefnId: defnId,
	}

	id := fmt.Sprintf("%v", defnId)
	if err := c.MetakvSet(DeleteDDLCommandTokenPath+id, commandToken); err != nil {
		return errors.New(fmt.Sprintf("Fail to delete index.  Internal Error = %v", err))
	}

	return nil
}

//
// Does token exist? Return true only if token exist and there is no error.
//
func DeleteCommandTokenExist(defnId c.IndexDefnId) (bool, error) {

	commandToken := &DeleteCommandToken{}
	id := fmt.Sprintf("%v", defnId)
	return c.MetakvGet(DeleteDDLCommandTokenPath+id, commandToken)
}

//
// Unmarshall
//
func UnmarshallDeleteCommandToken(data []byte) (*DeleteCommandToken, error) {

	r := new(DeleteCommandToken)
	if err := json.Unmarshal(data, r); err != nil {
		return nil, err
	}

	return r, nil
}

func MarshallDeleteCommandToken(r *DeleteCommandToken) ([]byte, error) {

	buf, err := json.Marshal(&r)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

//////////////////////////////////////////////////////////////
// Version Management
//////////////////////////////////////////////////////////////

//
// Generate a token to metakv for indexer version
//
func PostIndexerVersionToken(version uint64) error {

	token := &IndexerVersionToken{
		Version: version,
	}

	if err := c.MetakvSet(IndexerVersionTokenPath, token); err != nil {
		logging.Errorf("Fail to post indexer version to metakv.  Internal Error = %v", err)
		return err
	}

	return nil
}

//
// Does token exist? Return true only if token exist and there is no error.
//
func GetIndexerVersionToken() (uint64, error) {

	token := &IndexerVersionToken{}
	found, err := c.MetakvGet(IndexerVersionTokenPath, token)
	if err != nil {
		logging.Errorf("Fail to get indexer version from metakv.  Internal Error = %v", err)
		return 0, err
	}

	if !found {
		return 0, nil
	}

	return token.Version, nil
}

//
// Unmarshall
//
func UnmarshallIndexerVersionToken(data []byte) (*IndexerVersionToken, error) {

	r := new(IndexerVersionToken)
	if err := json.Unmarshal(data, r); err != nil {
		return nil, err
	}

	return r, nil
}

func MarshallIndexerVersionToken(r *IndexerVersionToken) ([]byte, error) {

	buf, err := json.Marshal(&r)
	if err != nil {
		return nil, err
	}

	return buf, nil
}
