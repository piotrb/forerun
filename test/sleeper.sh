# set -me

echo "sleeper starting: $$"

# function _bla {
#   exit 20
# }

# trap _bla INT

exec python inner.py
exit 22
