import App from './App.svelte';
import indexes from './markdown/indexes.md'
import documents from './markdown/documents.md'
import searching from './markdown/searching.md'
import deployment from './markdown/deployment.md'
import security from './markdown/security.md'
import access_keys from './markdown/access_keys.md'
import cloud_functions from './markdown/cloud_functions.md'
import client_libs_and_sdks from './markdown/client_libs_and_sdks.md'

const topics = [
  {
    subject: 'CONCEPTS',
    links: [
      {text: 'Indexes', id: 'indexes', markdown: indexes},
      {text: 'Documents', id: 'documents', markdown: documents},
      {text: 'Searching', id: 'searching', markdown: searching},
    ],
  },
  {
    subject: 'GUIDES',
    links: [
      {text: 'Deployment', id: 'deployment', markdown: deployment},
      {text: 'Security', id: 'security', markdown: security},
      {text: 'Access keys', id: 'access_keys', markdown: access_keys},
    ],
  },
  {
    subject: 'SOLUTIONS',
    links: [
      {text: 'Cloud Functions', id: 'cloud_functions', markdown: cloud_functions},
    ],
  },
  {
    subject: 'DEVELOPER',
    links: [
      {text: 'Client libs and SDKs', id: 'client_libs_and_sdks', markdown: client_libs_and_sdks},
    ],
  }
]

const app = new App({
	target: document.body,
	props: {
	  topics
  }
});

export default app;