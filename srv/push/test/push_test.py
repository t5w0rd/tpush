#!/usr/bin/env python

import asyncio
import websockets
import json
import sys
from typing import List

seq = 0

def genseq():
    global seq
    seq += 1
    return seq


def genreq(cmd: str, immed: bool, data: any) -> bytes:
    req = {
        'cmd': cmd,
        'seq': genseq(),
        'immed': immed,
        'data': data
    }
    req = [req,]
    bs = json.dumps(req).encode()
    return bs


async def login(address, uid) -> websockets.client.Connect:
    conn = await websockets.connect(address)
    req = {
        'uid': uid
    }

    s = genreq('login', True, req)
    #print('send:', s)
    await conn.send(s)
    return conn

recvcount = 0

async def recv_loop(conn: websockets.client.Connect):
    global recvcount
    while True:
        try:
            s = await conn.recv()
            rsps = json.loads(s)
            for rsp in rsps:
                #print(rsp)
                cmd = rsp['cmd']
                recvcount = recvcount + 1 
                print('recvcount:{0}'.format(recvcount), end="\r")
        except:
            conn.close()
            break


async def create_client_task(address: str, uid: int):
    conn = await login(address, uid)
    await recv_loop(conn)


async def main():
    import optparse
    op = optparse.OptionParser()
    op.add_option('-a', '--address', action='store', dest='address', help='Server address')
    op.add_option('-c', '--count', action='store', dest='count', help='Client count')
    op.add_option('-u', '--uid', action='store', dest='uid', help='User indentify')
    opts, args = op.parse_args()

    address = opts.address
    count = int(opts.count)
    uid = int(opts.uid)
    tasks: List = []
    for i in range(count):
        tasks.append(create_client_task(address, uid))

    #print('{0} logged in complete'.format(count))
    await asyncio.gather(*tasks)


if __name__ == '__main__':
    asyncio.run(main())

