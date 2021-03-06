package heis

import(
   "driver"
   "time"  
   "math"
  // "fmt"
)

func HeisInit()(int, int, int){
   direction := 0;
   for driver.GetFloor() == -1 {
      driver.SetSpeed(-300)
   }
   driver.SetSpeed(0)
   current_floor := driver.GetFloor()
   destination := -1
   return direction, current_floor, destination
}

func Heis(order_list chan driver.Data, command_list chan driver.Data, cost chan driver.Data, remove_order chan driver.Data, remove_command chan driver.Data, elevator_number chan int){ 
   //Initialize variables. 
   direction, current_floor, destination := HeisInit()
   var (
   		remove_orders 		driver.Data
   		remove_commands 	driver.Data
   		cost_copy 			driver.Data
   )
   order_list_copy 		:= driver.DataInit()
   command_list_copy 	:= driver.DataInit()
   elevator_nr 			:= -1
   
   //Starts a gorutine for continously reading relevant channels. 
   go func(){
      for{
         select{
         case data := <- order_list:
            order_list_copy = data
         case data := <- command_list:
            command_list_copy = data
         case data := <- elevator_number:
         	elevator_nr = data
         }
         time.Sleep(1*time.Millisecond)
      }
   }()
   
   for{
      
      destination = getDestination(direction, current_floor, order_list_copy.Array, command_list_copy.Array, elevator_nr)
      direction = getDirection(destination, current_floor)
      driver.SetSpeed(direction*300)
      
      //no point updating cost unless the elevator has discovered a new floor
      if(driver.GetFloor() != -1){
      	cost_copy.Array = costFunction(current_floor, direction, destination, order_list_copy.Array, elevator_nr)
      	cost <- cost_copy
      }
        
      for(destination != -1){
         destination = getDestination(direction, current_floor, order_list_copy.Array, command_list_copy.Array, elevator_nr)
         floor := driver.GetFloor() 
         if(floor != -1){
            current_floor = floor
         }
         if(driver.GetFloor() != -1){
      		cost_copy.Array = costFunction(current_floor, direction, destination, order_list_copy.Array, elevator_nr)
      		cost <- cost_copy
      	 }	
      	 
      	 
         //If sentence with the requirements for a stop. 
         if( (current_floor == driver.GetFloor()) && ((direction==-1 && order_list_copy.Array[2*current_floor]==elevator_nr) || (direction==1 && order_list_copy.Array[2*current_floor+1]==elevator_nr) || command_list_copy.Array[current_floor] == 1 || (destination == current_floor))){
        
            //stopping the elevator
            driver.SetSpeed(-1* direction*300)
            time.Sleep(15*time.Millisecond)
            driver.SetSpeed(0)
          	
          	//Sending what orders have been accomplished.
          	remove_orders.Array, remove_commands.Array = removeOrders(current_floor, direction, destination)
            remove_order <- remove_orders
            remove_command <- remove_commands
            
            //opening/closing doors 
            driver.SetDoorLamp(1)
            time.Sleep(3*time.Second)
            driver.SetDoorLamp(0)
            
            
            //resets the destination if it is the current floor.
            if current_floor == destination{
               destination = -1
            } 
            break
         }
         time.Sleep(1*time.Millisecond)
      }
      time.Sleep(1*time.Millisecond) 
   }
}







// Deciding what direction the destination lies in. 
func getDirection(destination int, current_floor int)(int){
   direction := 0
   if (destination == -1){
      return direction
   }else if(destination > current_floor){
      direction = 1  
   }else if(destination < current_floor){
      direction = -1
   }
   return direction
}

//Finding a destination and/or optimizing the destination. 
func getDestination(direction int, current_floor int, order_list [8]int, command_list [8]int, elevator_nr int)(int){
   var i int
   candidate := -1
   if(direction == 1){
      i = driver.N_FLOORS - 1
      for(i >= current_floor){
         if (order_list[i*2+1] == elevator_nr || command_list[i] == 1){
            if i > candidate{
            	candidate =  i
            }
         }else if(order_list[i*2] == elevator_nr){
         	if i > candidate{
         		candidate = i
         	}
         }
         i -= 1
      }
      return candidate
   }else if (direction == -1){
      i = 0
      for(i <= current_floor){
         if (order_list[i*2] == elevator_nr || command_list[i] == 1){
            return i
         }else if (order_list[i*2+1] == elevator_nr){
         	return i
         }
         i += 1
      }
      return -1
   }else{
      i = 0
      for(i < driver.N_FLOORS){
         if (order_list[i*2] == elevator_nr || order_list[i*2+1] == elevator_nr || command_list[i] == 1){
            return i
         }
         i += 1
      }
      return -1         
   }

}

// generating cost array
func costFunction(current_floor int,direction int, destination int, order_list [8]int, elevator_nr int)([8]int){
   i := 0
   var cost [driver.N_FLOORS*2]int
   for i<driver.N_FLOORS*2{
      if (direction == 0){
         cost[i] = int(math.Abs(float64(i/2 - current_floor)))
      }else if(direction == 1){
         if(i%2 == 1 && i/2 > current_floor){
            cost[i] = i/2 - current_floor - 1
         }else if (i%2 == 1 && i/2 <= current_floor || i%2 == 0){
            cost[i] = int (math.Abs(float64(i/2 - destination)) + math.Abs(float64(destination - current_floor - 1)))
         	if (destination != 3){
         		cost[i] += 1
         	}
         }else{
            cost[i] = 6
         }
      }else{
         if(i%2 == 0 && i/2 < current_floor){
            cost[i] =  current_floor - i/2 - 1
         }else if (i%2 == 0 && i/2 >= current_floor || i%2 ==  1){
            cost[i] = int(math.Abs(float64(i/2 - destination)) + math.Abs(float64(current_floor - destination - 1)))
         	if (destination != 0){
         		cost[i] += 1
         	}
         }else{
            cost[i] = 6
         }
      }
      i += 1
   }
   i = 0
   
   for i<driver.N_FLOORS*2{
      if order_list[i] == elevator_nr{
         cost[i] = 0
      }
      i += 1
   }
   return cost
}


// Making two arrays signalizing what orders have been completed
func removeOrders(current_floor int,direction int, destination int)([8]int,[8]int){
   remove_order := [8]int{0,0,0,0,0,0,0,0}
   remove_command := [8]int{0,0,0,0,0,0,0,0}
   i := 0
   for (i < driver.N_FLOORS){
      if (current_floor == i){
         remove_command[i] = 1
         if (destination == i){
         	remove_command[i] = 1
         	if (direction == -1 || i == driver.N_FLOORS-1){
         		remove_order[i*2] = 1
         	}else if (direction == 1 || i == 0){
         		remove_order[i*2+1] = 1
         	}else{
         		remove_order[i*2] = 1
         		remove_order[i*2+1] = 1
         	}
         	
         }else if (direction == 1){
            remove_order[i*2+1] = 1
         } else if (direction == -1){
            remove_order[i*2] = 1
         } 
      }
      i +=1
   }
   return remove_order, remove_command
}

