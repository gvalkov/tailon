#!/usr/bin/env python3

'''Writes log-file-looking lines to files.'''

import os
import time
import argparse
import asyncio
from random import choice, randint, seed
from datetime import datetime as dt


agents = (
    'Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.31 (KHTML, like Gecko) Chrome/26.0.1410.64 Safari/537.31',
    'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_8_3) AppleWebKit/537.31 (KHTML, like Gecko) Chrome/26.0.1410.65 Safari/537.31',
    'Mozilla/5.0 (Windows NT 6.1; WOW64; rv:20.0) Gecko/20100101 Firefox/20.0',
    'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_8_3) AppleWebKit/536.29.13 (KHTML, like Gecko) Version/6.0.4 Safari/536.29.13',
    'Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.31 (KHTML, like Gecko) Chrome/26.0.1410.64 Safari/537.31',
    'Mozilla/5.0 (Windows NT 6.2; WOW64) AppleWebKit/537.31 (KHTML, like Gecko) Chrome/26.0.1410.64 Safari/537.31',
    'Opera/9.80 (Windows NT 6.1; WOW64) Presto/2.12.388 Version/12.15',
)

paths = (
    '/js/app/app.js',
    '/js/app/router.js',
    '/js/app/model/User.js',
    '/js/app/model/Chat.js',
    '/js/app/model/ChatMessage.js',
    '/js/app/view/NavView.js',
    '/js/app/view/AppView.js',
    '/js/app/view/TrackView.js',
    '/js/app/view/ChatView.js',
    '/js/app/view/HomeView.js',
    '/js/app/view/DiscoverView.js',
    '/js/app/view/SignupView.js',
    '/js/app/view/CreateRoomView.js',
    '/js/app/view/ListenersView.j',
    '/js/app/view/LoginView.js',
    '/index.html',
)

methods = 'POST', 'GET', 'HEAD'
codes = 304, 404, 300, 400, 200
logfmt = '[{now:%d/%b/%Y:%H:%M:%S %z}] "{method} {path} HTTP/1.1" {status} 0 "{agent}"\n'


def generate_lines():
    while True:
        yield logfmt.format(
            now=dt.now(),
            method=choice(methods),
            path=choice(paths),
            status=choice(codes),
            agent=choice(agents),
        )


async def writer(fn, gen, lock, rate=(1, 4), update_msec=(500, 1000)):
    while True:
        n = randint(*rate) if isinstance(rate, (tuple, list)) else rate
        s = randint(*update_msec) if isinstance(update_msec, (tuple, list)) else update_msec

        async with lock:
            with open(fn, 'a') as fh:
                for i in range(n):
                    fh.write(next(gen))
                fh.flush()

        await asyncio.sleep(s/1000)


async def truncater(fn, lock, truncate_msec=10000):
    while True:
        await asyncio.sleep(truncate_msec/1000)
        async with lock:
            fh = open(fn, 'w')
            fh.close()


async def logwriter(args):
    gen = generate_lines()
    lock = asyncio.Lock()

    coros = []
    for fn in args.files:
        w = writer(fn, gen, lock=lock, rate=args.rate, update_msec=args.update_msec)
        coros.append(w)

        if not args.no_truncate:
            t = truncater(fn, lock=lock, truncate_msec=args.truncate_msec)
            coros.append(t)

    await asyncio.gather(*coros)


def main():
    def tuple_or_int(value):
        if ',' in value:
            return [int(i) for i in value.split(',')]
        else:
            return int(value)

    parser = argparse.ArgumentParser()
    arg = parser.add_argument
    arg('--update-msec',   default=1000,  metavar='msec', type=tuple_or_int)
    arg('--truncate-msec', default=10000, metavar='msec', type=tuple_or_int)
    arg('--no-truncate', action='store_false')
    arg('--rate', default=1, metavar='msec', type=tuple_or_int)
    arg('--seed', default=str(time.time()))
    arg('files', nargs=argparse.REMAINDER)

    args = parser.parse_args()
    args.files = [os.path.abspath(fn) for fn in args.files]

    print('using random seed: %s' % args.seed)
    seed(args.seed)

    loop = asyncio.get_event_loop()
    loop.run_until_complete(logwriter(args))
    loop.close()


if __name__ == '__main__':
    main()
