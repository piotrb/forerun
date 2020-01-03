import time
import signal, os
import atexit

def handler(signum, frame):
  print('Signal handler called with signal', signum)
  exit(1)

signal.signal(signal.SIGTERM, handler)
signal.signal(signal.SIGINT, handler)

def all_done():
    print 'all_done()'

atexit.register(all_done)

for i in range(5):
  print(i)
  time.sleep(1)
