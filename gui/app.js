console.log("I'm alive !")

let root, wsnd;

fetch('http://' + location.host + '/messages').then(response => response.json()).then(json => root = protobuf.Root.fromJSON(json))

const ws = new WebSocket('ws://' + location.host + '/ws')
ws.onopen = () => {
    ws.binaryType = 'arraybuffer'
    console.log('Opened')
}
ws.onmessage = message => {
    wsnd = root.lookupType('message.WSDeviceStatus').decode(new Uint8Array(message.data.slice(1)));
    console.log(wsnd.toJSON())
}