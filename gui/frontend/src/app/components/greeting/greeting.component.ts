import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { WailsService } from '../../services/wails.service';

@Component({
  selector: 'app-greeting',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './greeting.component.html',
  styleUrl: './greeting.component.css'
})
export class GreetingComponent {
  name = '';
  result = 'Please enter your name below 👇';

  constructor(private wailsService: WailsService) {}

  async greet(): Promise<void> {
    // Check if the input is empty
    if (this.name === '') return;

    try {
      // Call App.Greet(name)
      const result = await this.wailsService.greet(this.name);
      // Update result with data back from App.Greet()
      this.result = result;
    } catch (err) {
      console.error(err);
      this.result = 'Error greeting user';
    }
  }
}
