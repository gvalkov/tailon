import pytest
import psutil
import aiohttp

import time
import subprocess


@pytest.fixture(scope='session')
def unix_socket(tmpdir_factory):
    return str(tmpdir_factory.mktemp('tailon') / 'httpd.sock')


@pytest.fixture(scope='session')
def server(request, unix_socket):
    def inner(*args):
        cmd = ['./tailon', '-b', unix_socket, *args]

        proc = subprocess.Popen(cmd, cwd='../')

        # Convenience methods.
        proc.sock = unix_socket
        proc.get_children = lambda: get_child_procs(proc.pid)

        # TODO: Read stdout to determine when server is ready.
        time.sleep(0.20)

        request.addfinalizer(proc.terminate)
        return proc

    return inner


@pytest.fixture()
async def client(request, unix_socket, event_loop):
    conn = aiohttp.UnixConnector(path=unix_socket)
    session = aiohttp.ClientSession(connector=conn)

    def close():
        async def aclose():
            await session.close()
        event_loop.run_until_complete(aclose())
    request.addfinalizer(close)

    return session


def get_child_procs(pid):
        procs = psutil.process_iter(attrs=['pid', 'ppid'])
        procs = [i for i in procs if i.ppid() == pid]
        return procs
