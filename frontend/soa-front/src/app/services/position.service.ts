import { Injectable } from '@angular/core';

export interface TouristPosition {
  touristId: string;
  latitude: number;
  longitude: number;
}

@Injectable({
  providedIn: 'root'
})
export class PositionService {
  private storageKey = 'tourist-position';

  savePosition(position: TouristPosition): void {
    localStorage.setItem(this.storageKey, JSON.stringify(position));
  }

  getPosition(): TouristPosition | null {
    const saved = localStorage.getItem(this.storageKey);
    return saved ? JSON.parse(saved) : null;
  }
}