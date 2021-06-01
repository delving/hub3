if [ -d "cidoc-crm" ]; then
  rm -r cidoc-crm
fi
if [ -f "cidoc-crm.zip" ]; then
  rm cidoc-crm.zip
fi

mkdir cidoc-crm
cp -r server/* cidoc-crm
if [ -d "cidoc-crm/models" ]; then
  rm -r cidoc-crm/models
fi

mkdir cidoc-crm/models
cp -r public cidoc-crm
zip -r cidoc-crm.zip cidoc-crm
rm -r cidoc-crm
echo Done!