import App from './App.svelte';

const config = {
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
}

const app = new App({
  target: document.body,
  props: {
    config
  }
});

export default app;