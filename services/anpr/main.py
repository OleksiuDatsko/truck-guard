import time
import logging
import numpy as np
import cv2
import torch
import torchvision
from fastapi import FastAPI, File, UploadFile
from fastapi.responses import JSONResponse
from nomeroff_net import pipeline

import torch.serialization
old_load = torch.load
def new_load(*args, **kwargs):
    kwargs['weights_only'] = False
    return old_load(*args, **kwargs)
torch.load = new_load

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger("anpr-service")

app = FastAPI(title="TruckGuard ANPR Service")

logger.info("Loading Nomeroff Net pipeline...")
nmr_pipeline = pipeline("number_plate_detection_and_reading", image_loader=None)
logger.info("Pipeline loaded successfully!")

@app.get("/health")
async def health():
    return {"status": "ok", "device": "cuda" if torch.cuda.is_available() else "cpu"}

@app.post("/recognize")
async def recognize_plate(file: UploadFile = File(...)):
    start_time = time.time()
    try:
        contents = await file.read()
        nparr = np.frombuffer(contents, np.uint8)
        img = cv2.imdecode(nparr, cv2.IMREAD_COLOR)

        if img is None:
            return JSONResponse(content={"error": "Invalid image"}, status_code=400)

        result = nmr_pipeline([img])
        response_data = []
        
        if result and len(result) > 0:
            image_outputs = result[0]
            
            item_1 = image_outputs[-1]
            item_2 = image_outputs[-2]
            
            if len(item_1) > 0 and isinstance(item_1[0], str):
                texts = item_1
                scores = item_2
            else:
                texts = item_2
                scores = item_1

            for i in range(len(texts)):
                text_val = str(texts[i]).upper().replace(" ", "")
                
                try:
                    raw_score = scores[i]
                    if isinstance(raw_score, (list, np.ndarray)):
                        score_val = float(raw_score[0])
                    else:
                        score_val = float(raw_score)
                except (IndexError, TypeError, ValueError):
                    score_val = 1.0

                response_data.append({
                    "plate": text_val,
                    "confidence": round(score_val, 4)
                })

        if not response_data:
            return {"found": False, "message": "No plates detected"}

        response_data.sort(key=lambda x: x["confidence"], reverse=True)

        return {
            "found": True,
            "plates": response_data,
            "meta": {"processing_time": time.time() - start_time, "count": len(response_data)}
        }
    except Exception as e:
        logger.error(f"Prediction error: {e}", exc_info=True)
        return JSONResponse(content={"error": str(e)}, status_code=500)
