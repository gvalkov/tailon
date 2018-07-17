import json
import asyncio
from urllib.parse import urljoin

import pytest
import aiohttp


@pytest.mark.asyncio
@pytest.mark.parametrize("root", ['/', 'tailon/', 'tailon/tailon/'])
async def test_relativeroot(server, client, root):
    proc = server('--relative-root', root, 'testdata/ex1/var/log/1.log')
    for path in '', 'ws', 'vfs/dist/main.js':
        url = urljoin('http://localhost', root + path)
        res = await client.get(url)
        assert res.status == 200

    url = urljoin('http://localhost', root + 'vfs/dist/non-existant')
    res = await client.get(url)
    out = await res.text()
    assert res.status == 404, out


@pytest.mark.asyncio
async def test_sockjs_list(server, client):
    proc = server('testdata/ex1/var/log/1.log')
    async with client.ws_connect('http://localhost/ws/0/0/websocket') as ws:
        await ws.send_json(['list'])
        res = await get_sockjs_response(ws)

        assert '__default__' in res
        assert res['__default__'][0]['path'] == 'testdata/ex1/var/log/1.log'


@pytest.mark.asyncio
async def test_sockjs_frontendmessage(server, client):
    proc = server('testdata/ex1/var/log/1.log', 'testdata/ex1/var/log/2.log')

    msg_tail = '''{"command":"tail","script":null,"entry":{"path":"testdata/ex1/var/log/1.log","alias":"/tmp/t1","size":14342,"mtime":"2018-07-14T15:07:33.524768369+02:00","exists":true},"nlines":10}'''

    msg_grep = '''{"command":"grep","script":".*","entry":{"path":"testdata/ex1/var/log/1.log","alias":"/tmp/t1","size":14342,"mtime":"2018-07-14T15:07:33.524768369+02:00","exists":true},"nlines":10}'''

    async with client.ws_connect('http://localhost/ws/0/0/websocket') as ws:
        await ws.send_json([msg_tail])
        await asyncio.sleep(0.1)

        assert len(proc.get_children()) == 1

        await ws.send_json([msg_grep])
        await asyncio.sleep(0.1)
        assert len(proc.get_children()) == 2

        await ws.send_json([msg_tail])
        await asyncio.sleep(0.1)
        assert len(proc.get_children()) == 1


async def get_sockjs_response(ws):
        assert (await ws.receive()).data == 'o'
        res = (await ws.receive()).data[1:]
        res = json.loads(res)
        return json.loads(res[0])
