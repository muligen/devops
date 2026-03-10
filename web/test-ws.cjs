const WebSocket = require('ws');

async function testWebSocket() {
  // Get token
  const token = await fetch('http://localhost:8080/api/v1/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username: 'admin', password: 'admin123' })
  })
    .then(r => r.json())
    .then(d => d.data.access_token);

  console.log('Token obtained');

  // Test direct connection
  const ws = new WebSocket(`ws://localhost:8080/api/v1/ws/dashboard?token=${token}`);

  ws.on('open', () => {
    console.log('Direct WebSocket connected!');
  });

  ws.on('message', (data) => {
    console.log('Direct message:', data.toString().slice(0, 200));
  });

  ws.on('error', (err) => {
    console.log('Direct WebSocket error:', err.message);
  });

  // Wait a bit then test through nginx
  setTimeout(() => {
    ws.close();

    console.log('\nTesting through nginx proxy...');
    const ws2 = new WebSocket(`ws://localhost:3000/api/v1/ws/dashboard?token=${token}`);

    ws2.on('open', () => {
      console.log('Nginx WebSocket connected!');
    });

    ws2.on('message', (data) => {
      console.log('Nginx message:', data.toString().slice(0, 200));
    });

    ws2.on('error', (err) => {
      console.log('Nginx WebSocket error:', err.message);
    });

    setTimeout(() => {
      ws2.close();
      process.exit(0);
    }, 3000);
  }, 3000);
}

testWebSocket();
