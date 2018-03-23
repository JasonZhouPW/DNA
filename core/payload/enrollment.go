/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package payload

import (
	"github.com/Ontology/crypto"
	"io"
)

type Enrollment struct {
	PublicKey *crypto.PubKey
}

func (e *Enrollment) Serialize(w io.Writer) error {
	if err := e.PublicKey.Serialize(w); err != nil {
		return err
	}
	return nil
}

func (e *Enrollment) Deserialize(r io.Reader) error {
	pk := new(crypto.PubKey)
	if err := pk.DeSerialize(r); err != nil {
		return err
	}
	e.PublicKey = pk
	return nil
}