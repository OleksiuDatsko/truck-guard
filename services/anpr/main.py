import sys
import os
import uvicorn
import torch
import functools
import numpy as np
import cv2

_original_torch_load = torch.load

@functools.wraps(_original_torch_load)
def unsafe_torch_load(*args, **kwargs):
    if 'weights_only' not in kwargs:
        kwargs['weights_only'] = False
    return _original_torch_load(*args, **kwargs)

torch.load = unsafe_torch_load

from fastapi import FastAPI, File, UploadFile
from fastapi.responses import JSONResponse
from nomeroff_net import pipeline

app = FastAPI(title="Nomeroff Net API")

print("Loading Nomeroff Net pipeline...")
number_plate_detection_and_reading = pipeline("number_plate_detection_and_reading", image_loader=None)
print("Pipeline loaded!")

@app.post("/recognize")
async def recognize_plate(file: UploadFile = File(...)):
    try:
        contents = await file.read()
        nparr = np.frombuffer(contents, np.uint8)
        img = cv2.imdecode(nparr, cv2.IMREAD_COLOR)

        if img is None:
             return JSONResponse(content={"error": "Could not decode image"}, status_code=400)

        result = number_plate_detection_and_reading([img])
        
        
        predicted_texts = []
        
        if len(result) > 0:
            image_data = result[0]
            texts_list = image_data[-1]
            predicted_texts = texts_list

        if not predicted_texts:
            return {"found": False, "message": "No license plate detected"}

        response_data = []
        for text in predicted_texts:
            response_data.append({
                "plate": str(text),
            })

        return {
            "found": True,
            "plates": response_data
        }

    except Exception as e:
        import traceback
        traceback.print_exc()
        return JSONResponse(content={"error": str(e)}, status_code=500)

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8000)