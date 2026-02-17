// Unstable Build LLC ("COMPANY") CONFIDENTIAL
//
// Unpublished Copyright (c) 2017-2026 Unstable Build, All Rights Reserved.
//
// NOTICE: All information contained herein is, and remains the property of COMPANY.
// The intellectual and technical concepts contained herein are proprietary to
// COMPANY and may be covered by U.S. and Foreign Patents, patents in process,
// and are protected by trade secret or copyright law. Dissemination of this information
// or reproduction of this material is strictly forbidden unless prior written permission
// is obtained from COMPANY. Access to the source code contained herein is hereby
// forbidden to anyone except current COMPANY employees, managers or contractors who
// have executed Confidentiality and Non-disclosure agreements explicitly covering such access.
//
// The copyright notice above does not evidence any actual or intended publication or
// disclosure of this source code, which includes information that is confidential and/or
// proprietary, and is a trade secret, of COMPANY. ANY REPRODUCTION, MODIFICATION,
// DISTRIBUTION, PUBLIC  PERFORMANCE, OR PUBLIC DISPLAY OF OR THROUGH USE OF THIS SOURCE CODE
// WITHOUT  THE EXPRESS WRITTEN CONSENT OF COMPANY IS STRICTLY PROHIBITED, AND IN
// VIOLATION OF APPLICABLE LAWS AND INTERNATIONAL TREATIES. THE RECEIPT OR POSSESSION OF
// THIS SOURCE CODE AND/OR RELATED INFORMATION DOES NOT CONVEY OR IMPLY ANY RIGHTS TO
// REPRODUCE, DISCLOSE OR DISTRIBUTE ITS CONTENTS, OR TO MANUFACTURE, USE, OR SELL
// ANYTHING THAT IT MAY DESCRIBE, IN WHOLE OR IN PART.

package main

import "fmt"

// Greeter greets people.
type Greeter struct {
	Name string
}

// Greet returns a greeting message.
func (g *Greeter) Greet() string {
	return fmt.Sprintf("Hello, %s!", g.Name)
}

// Add adds two integers.
func Add(a, b int) int {
	return a + b
}

func main() {
	g := &Greeter{Name: "World"}
	fmt.Println(g.Greet())
	fmt.Println(Add(1, 2))
}

// Speaker is an interface for things that can speak.
// See also [Add] for arithmetic operations.
type Speaker interface {
	Speak() string
}

// Robot is a mechanical speaker.
type Robot struct {
	ID int
}

// Speak returns a robotic greeting.
func (r *Robot) Speak() string {
	return fmt.Sprintf("Beep boop, I am robot %d", r.ID)
}
