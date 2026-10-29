#!/usr/bin/env bash
#
# How much of the index carries meta.hasDigitalObject (#3052).
#
# The flag is only written when a record is indexed, and we chose to let the
# index heal rather than force a full reindex. Until healing is far enough
# along, a "met media / zonder media" facet lies -- and lies worst on the half
# the customer asked for, because a record without the field falls into neither
# bucket. "Zonder media" then looks nearly empty while hundreds of thousands of
# records belong in it.
#
# So this is the gate on switching the facet on, not a progress bar.
#
# Usage: heal-coverage.sh [ssh-host] [index]
set -uo pipefail
HOST="${1:-root@ingestion.brabantcloud.acpt.delving.io}"
INDEX="${2:-brabantcloudv2}"
DREMPEL=95

count() {
  ssh -o ConnectTimeout=25 "$HOST" \
    "curl -s --max-time 60 -H 'Content-Type: application/json' \
      'http://localhost:9200/$INDEX/_count' -d '$1'" 2>/dev/null |
    python3 -c 'import sys,json;print(json.load(sys.stdin)["count"])' 2>/dev/null
}

totaal=$(count '{"query":{"match_all":{}}}')
geheeld=$(count '{"query":{"exists":{"field":"meta.hasDigitalObject"}}}')
met=$(count '{"query":{"term":{"meta.hasDigitalObject":true}}}')
zonder=$(count '{"query":{"term":{"meta.hasDigitalObject":false}}}')

if [ -z "${totaal:-}" ] || [ "${totaal:-0}" = "0" ]; then
  echo "kon de index niet lezen: $HOST / $INDEX" >&2
  exit 2
fi

pct=$(python3 -c "print(f'{100*$geheeld/$totaal:.1f}')")

printf 'index      %12d records\n' "$totaal"
printf 'met vlag   %12d  (%s%%)\n' "$geheeld" "$pct"
printf '  true     %12d\n' "$met"
printf '  false    %12d\n' "$zonder"
printf 'zonder vlag%12d  -- in geen van beide facet-emmers\n' "$((totaal - geheeld))"
echo

if python3 -c "import sys; sys.exit(0 if $pct >= $DREMPEL else 1)"; then
  echo "BOVEN DE DREMPEL ($DREMPEL%): het facet vertelt de waarheid, aanzetten mag."
  exit 0
fi

echo "ONDER DE DREMPEL ($DREMPEL%): facet nog niet aanzetten."
echo "Zonder-media zou $((totaal - geheeld)) records missen die er wel in horen."
exit 1
