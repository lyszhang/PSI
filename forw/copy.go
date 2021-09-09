/**
 * @Author: lyszhang
 * @Email: ericlyszhang@gmail.com
 * @Date: 2021/8/22 5:10 PM
 */

package main

import (
	"io"
	"sync"
)

var bufpool *sync.Pool

func init() {
	bufpool = &sync.Pool{}
	bufpool.New = func() interface{} {
		return make([]byte, 24*1024)
	}
}

func Copy(dst io.Writer, src io.Reader) (written int64, err error) {
	if wt, ok := src.(io.WriterTo); ok {
		return wt.WriteTo(dst)
	}
	if rt, ok := dst.(io.ReaderFrom); ok {
		return rt.ReadFrom(src)
	}

	buf := bufpool.Get().([]byte)
	defer bufpool.Put(buf)

	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			dst.Write(buf[0:nr])

			//nw, ew := dst.Write(buf[0:nr])
			//if nw > 0 {
			//	written += int64(nw)
			//}
			//if ew != nil {
			//	err = ew
			//	break
			//}
			//if nr != nw {
			//	err = io.ErrShortWrite
			//	break
			//}
		}
		if er == io.EOF {
			break
		}
		//if er != nil {
		//	err = er
		//	break
		//}
	}
	return written, err
}
