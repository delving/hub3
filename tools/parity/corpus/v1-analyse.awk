# Leest nginx-combined. Per /api/search-request:
#  - telt elke parameternaam
#  - bouwt een "signature": de gesorteerde set parameternamen
#  - bewaart per signature een handvol echte URL's als replay-voorbeeld
{
  if (!match($0, /"(GET|HEAD|POST)[^"]*"/)) next
  req = substr($0, RSTART+1, RLENGTH-2)
  split(req, a, " ")
  url = a[2]
  if (index(url, "/api/search") != 1) next
  totaal++

  qi = index(url, "?")
  if (qi == 0) { leeg++; next }
  qs = substr(url, qi+1)

  n = split(qs, paren, "&")
  delete namen
  m = 0
  for (i = 1; i <= n; i++) {
    p = paren[i]
    ei = index(p, "=")
    naam = (ei ? substr(p, 1, ei-1) : p)
    if (naam == "") continue
    # hqf[]/qf[] etc: array-haakjes weg, anders explodeert de signature
    gsub(/%5B%5D|\[\]/, "[]", naam)
    param[naam]++
    if (!(naam in namen)) { namen[naam] = 1; m++; lijst[m] = naam }
  }
  # sorteer de namen (kleine m, insertion sort volstaat)
  for (i = 2; i <= m; i++) {
    v = lijst[i]; j = i-1
    while (j > 0 && lijst[j] > v) { lijst[j+1] = lijst[j]; j-- }
    lijst[j+1] = v
  }
  sig = ""
  for (i = 1; i <= m; i++) sig = sig (i>1 ? "+" : "") lijst[i]
  sigtel[sig]++
  if (voorbeeldtel[sig] < 5) { voorbeeldtel[sig]++; voorbeeld[sig SUBSEP voorbeeldtel[sig]] = url }
}
END {
  print "# TOTAAL", totaal, "zonder-query", leeg+0 > UIT "/samenvatting.txt"
  for (p in param) print param[p], p > UIT "/parameters.txt"
  for (s in sigtel) {
    print sigtel[s], s > UIT "/signatures.txt"
    for (i = 1; i <= voorbeeldtel[s]; i++) print sigtel[s] "\t" s "\t" voorbeeld[s SUBSEP i] > UIT "/voorbeelden.tsv"
  }
}
