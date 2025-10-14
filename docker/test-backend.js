const express = require('express');
const app = express();
const port = 3000;

app.use(express.json());

// Simple echo endpoint
app.post('/chat', (req, res) => {
  console.log('Received query parameters:', JSON.stringify(req.query, null, 2));
  console.log('Received body request:', JSON.stringify(req.body, null, 2));
  
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

// Health check
app.get('/health', (req, res) => {
  res.json({ status: 'healthy', timestamp: new Date().toISOString() });
});

// Generic JSON endpoint
app.post('/api/:endpoint', (req, res) => {
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