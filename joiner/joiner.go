package joiner

import (
	"github.com/grafov/m3u8"
	"io/ioutil"
	"os"
	"sync"
	"time"
)

type Joiner struct {
	l      sync.Mutex
	playlist *m3u8.MediaPlaylist
	blocks map[uint]string
	file   *os.File
	name   string
}

func New(name string, pl *m3u8.MediaPlaylist) (*Joiner, error) {
	f, err := os.OpenFile(name, os.O_CREATE|os.O_TRUNC|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	joiner := &Joiner{
		playlist: pl,
		blocks: map[uint]string{},
		file:   f,
		name:   name,
	}

	return joiner, nil
}

//func (j *Joiner) Join(id int, block []byte) {
//	j.l.Lock()
//	j.blocks[id] = block
//	j.l.Unlock()
//}
func (j *Joiner) JoinFile(id uint, path string) {
	j.l.Lock()
	j.blocks[id] = path
	j.l.Unlock()
}
func (j *Joiner) Run(count uint) (string, error) {
	var index uint = 0
	//playlist, err := m3u8.NewMediaPlaylist(count, count)
	//j.playlist = playlist
	var err error
	var result string
	var content []byte
	var tsInfo os.FileInfo
	var offset int64 = 0
	for{
		if uint(len(j.blocks)) == count {
			for ;index < count; index++ {
				j.l.Lock()
				block := j.blocks[index]
				j.l.Unlock()
				if block != "" {
					tsInfo, err = os.Stat(block)
					limit := tsInfo.Size()
					//j.playlist.SetRange(s, offset)
					j.playlist.Segments[index].Limit = limit
					j.playlist.Segments[index].Offset = offset
					j.playlist.Segments[index].URI = j.name
					offset += limit
					content,  err = ioutil.ReadFile(block)
					if err != nil{
						return result, err
					}
					_, err = j.file.Write(content)
					if err != nil {
						return result, err
					}
					//j.l.Lock()
					//delete(j.blocks, index)
					//j.l.Unlock()
					//index++
				}
			}
			break
		} else {
			time.Sleep(time.Second * 2)
		}
	}

	result = j.playlist.String()
	j.playlist.Close()
	return result, j.file.Close()
}

func (j *Joiner) Name() string {
	return j.name
}
