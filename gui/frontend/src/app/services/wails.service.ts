import { Injectable } from '@angular/core';
import { Greet, GetGNSSGOVersion } from '../../wailsjs/go/main/App';

@Injectable({
  providedIn: 'root'
})
export class WailsService {
  constructor() { }

  /**
   * Greets the user with the provided name
   * @param name The name to greet
   * @returns A promise that resolves to the greeting message
   */
  async greet(name: string): Promise<string> {
    try {
      return await Greet(name);
    } catch (error) {
      console.error('Error greeting user:', error);
      throw error;
    }
  }

  /**
   * Gets the GNSSGO version from the backend
   * @returns A promise that resolves to the GNSSGO version
   */
  async getGNSSGOVersion(): Promise<string> {
    try {
      return await GetGNSSGOVersion();
    } catch (error) {
      console.error('Error getting GNSSGO version:', error);
      throw error;
    }
  }
}
