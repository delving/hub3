export let config = {
  facets: {
    'tree.hasDigitalObject': {
      order: 1,
      label: 'Bevat digitaal materiaal'
    },
    'ead-rdf_genreform': {
      order: 2,
      label: 'Soort materiaal'
    },
    'tree.mimeType': {
      order: 3,
      label: 'Digitaal bestandstype'
    }
  },
  urls:
    {
      ead: {
        path: ['archief', ':inventoryID']
      },
      cLevel: {
        path: ['archief', ':inventoryID', 'invnr', ':cLevelPath']
      },
      eadDescription: {
        path: ['archief', ':inventoryID', 'description'],
        query: [
          {key: 'q', value: ':query'}
        ]
      }
    }
};