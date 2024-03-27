import json
import datetime
from typing import Union, List
from fastapi import APIRouter
from fastapi import Request
from pydantic import BaseModel
import pandas as pd
from icecream import ic


router = APIRouter()



@router.post("/process")
async def process(request: Request):
    
    body_dict = await request.json()
    print(body_dict)
    
    summary = body_dict.get("issue", {}).get("fields", {}).get("summary", None)
    
    with open(f"temp_{datetime.datetime.now().strftime('%Y%m%d_%H%M%S')}.json", "w") as f:
        json.dump(body_dict, f)

    return {"summary": summary}




@router.post("/calculate_roof_price_per_quarter")
async def calculate_roof_price_per_quarter(request: Request):

    # price_column_name = "Negative imbalance price"
    price_column_name = "price"

    body_dict = await request.json()
    
    df = pd.json_normalize(body_dict.get("time_series_data").get("$values"))

    ev_comfort_charge_capacity_kwh = int(body_dict.get("ev_comfort_charge_capacity_kwh"))
    ev_max_charge_capacity_kwh = int(body_dict.get("ev_max_charge_capacity_kwh"))
    buffer = float(body_dict.get("buffer"))
    
    with open(f"data_request_body_jsons/temp_{datetime.datetime.now().strftime('%Y%m%d_%H%M%S')}.json", "w") as f:
        json.dump(body_dict, f)
    

    ev_pmax = 22  # kW
    ev_charged_per_hour = ev_pmax  # kWh
    cutoff_time = 17  # 17pm(ish)

    df = df.head(cutoff_time)


    ev_charging_hours_count_comfort = ev_comfort_charge_capacity_kwh / ev_charged_per_hour
    ev_charging_hours_count_max = ev_max_charge_capacity_kwh / ev_charged_per_hour
    total_hours_count = len(df)
    ic(ev_charging_hours_count_comfort, ev_charging_hours_count_max, total_hours_count)

    ic(ev_charging_hours_count_comfort / total_hours_count)

    percent_of_hours_needed_comfort = ev_charging_hours_count_comfort / total_hours_count
    percent_of_hours_needed_max = ev_charging_hours_count_max / total_hours_count
    ic(percent_of_hours_needed_comfort, percent_of_hours_needed_max)

    percent_of_hours_needed_comfort = percent_of_hours_needed_comfort * (1 + buffer)
    percent_of_hours_needed_max = percent_of_hours_needed_max * (1 + buffer)
    ic(percent_of_hours_needed_comfort, percent_of_hours_needed_max)

    df['is_in_lowest_hours_comfort'] = df[price_column_name] <= df[price_column_name].quantile(percent_of_hours_needed_comfort)
    df['is_in_lowest_hours_max'] = df[price_column_name] <= df[price_column_name].quantile(percent_of_hours_needed_max)

    pd.set_option('display.max_rows', None)
    print(df)

    highest_price_in_lowest_hours_comfort = df[df['is_in_lowest_hours_comfort']][price_column_name].max()
    highest_price_in_lowest_hours_max = df[df['is_in_lowest_hours_max']][price_column_name].max()

    try:
        last_hour_comfort = df[df['is_in_lowest_hours_comfort']].index[-1]
        last_hour_max = df[df['is_in_lowest_hours_max']].index[-1]
    except IndexError:
        last_hour_comfort = 0
        last_hour_max = 0

    # TODO room for improvement here, it could be somewhere between highest_price_in_lowest_quarters_comfort
    #      (if that's below 0) and highest_price_in_lowest_quarters_max
    if highest_price_in_lowest_hours_max > 0:
        highest_price_in_lowest_hours_max = highest_price_in_lowest_hours_comfort

    return {
        "roof_comfort": highest_price_in_lowest_hours_comfort,
        "roof_max": highest_price_in_lowest_hours_max,
        "last_hour_comfort": last_hour_comfort,
        "last_hour_max": last_hour_max
    }

