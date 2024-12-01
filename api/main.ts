// @deno-types="npm:@types/express@4.17.15"
import express from "npm:express@4.18.2";
import axios from "npm:axios"

const app = express();

app.post("/upload", async (req, res) => {
  const name = req?.body?.name;
  const buffer = Deno.readFileSync("data.csv")

  try{
    const response = await axios.post("http://localhost:3030/process",
      {
        file: Array.from(buffer),
        name
      },
      {
        headers:{
          "Content-Type": "application/json"
        }
      })

    if(response.status == 200 || response.status == 201){
      const {data} = response
      res.status(200).send(data)
    }else{
      res.status(400).send()
    }
  }catch(err){
    console.log(err)
    res.status(400).send()
  }
});


app.post("/subscribe/:id", async(req,res) => {
  const { id } = req.params;
  const STREAMING_SERVER_URL = `ws://localhost:3030/socket/${id}`;

  try {
    const socket = new WebSocket(STREAMING_SERVER_URL);

    socket.onopen = () => {
      console.log("Connected to the streaming server");
      socket.send(
        JSON.stringify({
          type: "subscribe",
        }),
      );
    };

    socket.onmessage = (event) => {
      console.log("Message received from server:", event.data);
      try {
        const message = JSON.parse(event.data);
        console.log("Parsed message:", message);
        socket.send(
          JSON.stringify({
            type: "next",
          }),
        );
      } catch (error) {
        console.error("Failed to parse message:", event.data,error);
      }
    };

    socket.onclose = (event) => {
      console.log(
        `Disconnected from the streaming server (Code: ${event.code})`,
      );
    };

    socket.onerror = (error) => {
      console.error("WebSocket error:", error);
    };
  } catch (error) {
    console.error("Error connecting to streaming server:", error);
  }
})


app.listen(8000,"localhost",()=>{
  console.log("listening on port ", 8000)
});
