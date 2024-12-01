'use server';

import { z } from 'zod';
import { MongoClient } from 'mongodb'
import { revalidatePath } from 'next/cache';
import { redirect } from 'next/navigation';
import axios from "axios"

const FormSchema = z.object({
  name: z.string().min(1, "Name is required"),
  file: z.instanceof(File)
});

const url = process.env.MONGO_URL;
if(!url) throw new Error("missing mongo url")
const dbName = process.env.MONGO_DATABASE;
if(!dbName) throw new Error("missing mongo database name")

const client = new MongoClient(url);
await client.connect();
const db = client.db(dbName);
const collection = db.collection('strategies');

export async function createStrategy(formData: FormData) {
  const parsedData = FormSchema.parse({
    name: formData.get("name"),
    file: formData.get("file"),
  });

  const { name, file } = parsedData;

  const buffer = await fileToBuffer(file);

  console.log(`Strategy Name: ${name}`);
  console.log(`File Buffer:`, buffer);
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
      const insertResult = await collection.insertOne({_id: data.id,url:data.url,name});
      console.log('Inserted documents =>', insertResult);
    }else{
      return {
        message: "Server Error: Failed to Upload the file."
      }
    }
  }catch(err){
    console.error(err)
    return {
      message: "Server Error: Failed to Upload the file."
    }
  }

  revalidatePath("/dashboard/strategies");
  redirect("/dashboard/strategies");
}

async function fileToBuffer(file: File): Promise<Buffer> {
  const arrayBuffer = await file.arrayBuffer();
  return Buffer.from(arrayBuffer);
}
