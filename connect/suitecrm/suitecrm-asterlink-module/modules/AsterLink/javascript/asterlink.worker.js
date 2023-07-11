// serfreeman1337 // 11.07.2023 //

let ports = [];
let currentUrl = undefined;
let stream = undefined;
let reconnectTimer = undefined;

let calls = [];

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
    const id = calls.findIndex(call => call.id == message.data.id)

    // Remember active calls so they can be send back to the new open tabs.
    if (message.show) {
      if (id == -1) {
        calls.push(message.data)
      } else {
        Object.assign(calls[id], message.data);
      }
    } else {
      calls.splice(id, 1);
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
  calls = []; // Reset calls on disconnect.e(id);

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
  for (const call of calls) {
    port.postMessage({ show: true, data: call });
  }
};