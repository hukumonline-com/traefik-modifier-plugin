const express = require('express');
const app = express();
const port = 3000;

app.use(express.json());

// Helper: log incoming request details (headers, body, query)
function logRequest(label, req) {
  console.log(`\n=== ${label} ===`);
  console.log('Method      :', req.method);
  console.log('Path        :', req.path);
  console.log('Query       :', JSON.stringify(req.query, null, 2));
  console.log('Headers     :', JSON.stringify(req.headers, null, 2));
  console.log('Body        :', JSON.stringify(req.body, null, 2));
}

// /chat — existing endpoint
app.post('/chat', (req, res) => {
  logRequest('/chat', req);

  const response = {
    id: Math.floor(Math.random() * 1000),
    answer: `You asked: "${req.body.question || req.body.message || 'nothing'}"`,
    timestamp: new Date().toISOString(),
    data: {
      processed: true,
      original_question: req.body.question,
      conversation_id: req.body.conversation_id
    },
    data_list: [1, 2, 3, 4, 5],
    data_map: { a: 1, b: 2, c: 3 },
    data_array_of_maps: [
      { key1: 'value1', key2: 'value2' },
      { key1: 'valueA', key2: 'valueB' }
    ]
  };

  console.log('Sending response:', JSON.stringify(response, null, 2));
  res.json(response);
});

// /v1/conversation — matches create-conversation-modifier route in dynamic.yaml
// Validates: .request.headers, .request.method, .request.url, .request.path, .request.modified.body.*
app.post('/v1/conversation', (req, res) => {
  logRequest('/v1/conversation', req);

  const response = {
    id: `conv-${Math.floor(Math.random() * 9000 + 1000)}`,
    answer: `Echo: "${req.body.message || req.body.question || 'nothing'}"`,
    session_id: req.body.session_id || null,   // should be populated by ModifierRequest
    meta_echo: req.body.meta || null,           // should contain method/url/path from plugin
    timestamp: new Date().toISOString(),
    // Mirror back key headers to verify they reached the backend
    received_headers: {
      authorization: req.headers['authorization'] || null,
      'x-api-key':   req.headers['x-api-key']   || null,
      'x-source':    req.headers['x-source']     || null
    }
  };

  console.log('Sending response:', JSON.stringify(response, null, 2));
  res.json(response);
});

// Health check
app.get('/health', (req, res) => {
  res.json({ status: 'healthy', timestamp: new Date().toISOString() });
});

// Generic JSON endpoint
app.post('/api/:endpoint', (req, res) => {
  logRequest(`/api/${req.params.endpoint}`, req);
  res.json({
    endpoint: req.params.endpoint,
    method: req.method,
    body: req.body,
    timestamp: new Date().toISOString()
  });
});

app.listen(port, () => {
  console.log(`Test backend running at http://localhost:${port}`);
});