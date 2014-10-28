// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package filestorage_test

import (
	"bytes"
	"io"
	"io/ioutil"

	"github.com/juju/testing"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils/filestorage"
)

var _ = gc.Suite(&WrapperSuite{})

type WrapperSuite struct {
	testing.IsolationSuite
	rawstor  *FakeRawFileStorage
	metastor *FakeMetadataStorage
	stor     filestorage.FileStorage
}

func (s *WrapperSuite) SetUpTest(c *gc.C) {
	s.IsolationSuite.SetUpTest(c)

	s.rawstor = &FakeRawFileStorage{}
	s.metastor = &FakeMetadataStorage{}
	s.stor = filestorage.NewFileStorage(s.metastor, s.rawstor)
}

func (s *WrapperSuite) metadata() filestorage.Metadata {
	meta := filestorage.NewMetadata(nil)
	meta.SetFile(10, "", "")
	return meta
}

func (s *WrapperSuite) setMeta() (string, filestorage.Metadata) {
	id := "<id>"
	meta := s.metadata()
	meta.SetID(id)
	s.metastor.meta = meta
	s.metastor.metaList = append(s.metastor.metaList, meta)
	return id, meta
}

func (s *WrapperSuite) setfile(data string) (string, filestorage.Metadata, io.ReadCloser) {
	id, meta := s.setMeta()
	file := ioutil.NopCloser(bytes.NewBufferString(data))
	s.rawstor.file = file
	meta.SetStored()
	return id, meta, file
}

func (s *WrapperSuite) TestFileStorageNewFileStorage(c *gc.C) {
	stor := filestorage.NewFileStorage(s.metastor, s.rawstor)

	c.Check(stor, gc.NotNil)
}

func (s *WrapperSuite) TestFileStorageMetadata(c *gc.C) {
	id, original := s.setMeta()
	meta, err := s.stor.Metadata(id)
	c.Check(err, gc.IsNil)

	c.Check(meta, gc.DeepEquals, original)
	s.metastor.Check(c, id, nil, "Metadata")
	s.rawstor.CheckNotUsed(c)
}

func (s *WrapperSuite) TestFileStorageGet(c *gc.C) {
	id, origmeta, origfile := s.setfile("spam")
	meta, file, err := s.stor.Get(id)
	c.Check(err, gc.IsNil)

	c.Check(meta, gc.Equals, origmeta)
	c.Check(file, gc.Equals, origfile)
}

func (s *WrapperSuite) TestFileStorageListEmpty(c *gc.C) {
	list, err := s.stor.List()
	c.Check(err, gc.IsNil)

	c.Check(list, gc.HasLen, 0)
}

func (s *WrapperSuite) TestFileStorageListOne(c *gc.C) {
	id, _ := s.setMeta()
	list, err := s.stor.List()
	c.Check(err, gc.IsNil)

	c.Check(list, gc.HasLen, 1)
	c.Assert(list[0], gc.NotNil)
	c.Check(list[0].ID(), gc.Equals, id)
}

func (s *WrapperSuite) TestFileStorageListTwo(c *gc.C) {
	id1, _ := s.setMeta()
	id2, _ := s.setMeta()
	list, err := s.stor.List()
	c.Check(err, gc.IsNil)

	c.Assert(list, gc.HasLen, 2)
	c.Assert(list[0], gc.NotNil)
	c.Assert(list[1], gc.NotNil)
	if list[0].ID() == id1 {
		c.Check(list[1].ID(), gc.Equals, id2)
	} else {
		c.Check(list[1].ID(), gc.Equals, id1)
	}
}

func (s *WrapperSuite) TestFileStorageAddMeta(c *gc.C) {
	s.metastor.id = "<spam>"

	meta := s.metadata()
	id, err := s.stor.Add(meta, nil)
	c.Assert(err, gc.IsNil)

	c.Check(id, gc.Equals, "<spam>")
	c.Check(meta.ID(), gc.Equals, "<spam>")
	s.metastor.Check(c, "", meta, "AddDoc")
	s.rawstor.CheckNotUsed(c)
}

func (s *WrapperSuite) TestFileStorageAddFile(c *gc.C) {
	s.metastor.id = "<spam>"

	var file *bytes.Buffer
	meta := s.metadata()
	id, err := s.stor.Add(meta, file)
	c.Check(err, gc.IsNil)

	c.Check(id, gc.Equals, "<spam>")
	c.Check(meta.ID(), gc.Equals, "<spam>")
	s.metastor.Check(c, "", meta, "AddDoc", "SetStored")
	s.rawstor.Check(c, id, file, 10, "AddFile")
}

func (s *WrapperSuite) TestFileStorageAddMetaOnly(c *gc.C) {
	id, original := s.setMeta()
	meta, err := s.stor.Metadata(id)
	c.Assert(err, gc.IsNil)

	c.Check(meta, gc.Equals, original)
	c.Check(meta.Stored(), gc.Equals, false)
}

func (s *WrapperSuite) TestFileStorageAddIDAlreadySet(c *gc.C) {
	original := s.metadata()
	original.SetID("eggs")
	_, err := s.stor.Add(original, nil)

	c.Check(err, gc.IsNil) // This should be handled at the lower level.
}

func (s *WrapperSuite) TestFileStorageSetFile(c *gc.C) {
	id, meta := s.setMeta()
	_, _, err := s.stor.Get(id)
	c.Assert(err, gc.NotNil)

	file := bytes.NewBufferString("spam")
	err = s.stor.SetFile(id, file)
	c.Assert(err, gc.IsNil)

	s.metastor.Check(c, id, meta, "Metadata", "Metadata", "SetStored")
	s.rawstor.Check(c, id, file, 10, "AddFile")
}

func (s *WrapperSuite) TestFileStorageRemove(c *gc.C) {
	id := "<spam>"
	err := s.stor.Remove(id)
	c.Assert(err, gc.IsNil)

	s.metastor.Check(c, id, nil, "RemoveDoc")
	s.rawstor.Check(c, id, nil, 0, "RemoveFile")
}
