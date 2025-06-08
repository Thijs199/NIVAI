from pydantic import BaseModel
from typing import Optional

class ProcessMatchRequest(BaseModel):
    tracking_data_path: str
    event_data_path: str
    match_id: Optional[str] = None

class BasicResponse(BaseModel):
    message: str
    match_id: Optional[str] = None

class StatusResponse(BaseModel):
    status: str
    match_id: Optional[str] = None
    message: Optional[str] = None

# More complex response models for specific stats endpoints can be added later
# if strictly typed outputs are required. For now, the API will return
# dictionaries/lists directly from the stats_calculator functions.
