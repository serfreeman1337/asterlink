// serfreeman1337 // 10.07.2023 //

let ports = [];
let currentUrl = undefined;
let stream = undefined;
let reconnectTimer = undefined;

let calls = new Map();

/**
 * Connect event stream.
 * @param {string} url Stream url.
 */
function connect(url) {
  if (!url) {
    disconnect();
    self.close();
    return;
  }

  if (stream) {
    if (url == currentUrl)  return;
    disconnect(); // New stream url is differs, disconnect previous before connecting a new one.
  }

  currentUrl = url;
  stream = new EventSource(url);
  reconnectTimer = undefined;

  stream.onmessage = e => {
    const message = JSON.parse(e.data);
    const id = message.data.id;

    // Remember active calls so they can be send back to the new open tabs.
    if (message.show) {
      calls.set(id, message.data);
    } else {
      calls.delete(id);
    }
    
    // Notify all tabs about call.
    for (const port of ports) {
      port.postMessage(message);
    }
  };

  stream.onerror = () => {
    disconnect();

    // Reconnect in 1s.
    currentUrl = '';
    reconnectTimer = setTimeout(() => connect(url), 1000);
  };
}

/**
 * Disconnect event stream.
 */
function disconnect() {
  for (const id of calls.keys()) { // Reset calls on disconnect.
    calls.delete(id);
  }

  for (const port of ports) { // Notify all tabs about disconnection.
    port.postMessage(false);
  }
  
  stream.close();
}

self.onconnect = e => {
  const port = e.ports[0];
  ports.push(port);



  port.onmessage = e => connect(e.data);

  // Send all active calls.
  for (const call of calls.values()) {
    port.postMessage({ show: true, data: call });
  }
};