---
title: Visitor Pattern
date: 2019-07-14 20:06:26
tags:
- visitor
- design pattern
categories:
- Design Pattern
---

There's a very simple routine in the book: ***Clean Code***, Chapter 6, page 96
``` java
public class Square {
	public Point topLeft;
	public double side;
}

public class Rectangle {
	public Point topLeft;
	public double height;
	public double width;
}

public class Circle {
	public Point center;
	public double radius;
}

public class Geometry {
	public final double PI = 3.141592653589793;
	public double area(Object shape) throws NoSuchShapeException
	{
		if (shape instanceof Square) {
			Square s = (Square)shape;
			return s.side * s.side;
		}
		else if (shape instanceof Rectangle) {
			Rectangle r = (Rectangle)shape;
			return r.height * r.width;
		}
		else if (shape instanceof Circle) {
			Circle c = (Circle)shape;
			return PI * c.radius * c.radius;
		}
		throw new NoSuchShapeException();
	}
}
```
As a OO programer, what a ugly code! That code are totally no object-oriented, and use such a if...else... structure to deal with different classes rather than using polymorphic.
However, uncle bob said at the follow:
> Consider what would happen if a perimeter() function were added to Geometry. The shape classes would be unaffected! Any other classes that depended upon the shapes would also be unaffected! On the other hand, if I add a new shape, I must change all the functions in Geometry to deal with it. Again, read that over. Notice that the two conditions are diametrically opposed. 

So the concept of Data Structure and Object are come out:
> Procedural code (code using data structures) makes it easy to add new functions without
changing the existing data structures. OO code, on the other hand, makes it easy to add
new classes without changing existing functions.

> Procedural code makes it hard to add new data structures because all the functions must
change. OO code makes it hard to add new functions because all the classes must change.

But again, as a OO programer, I can't take this, I want to deal with the scenario of change behavior more "elegantly".

Then Uncle bob jump out again and say: try Vistor pattern! (you can find it at the footnote in page 96)

### Define Vistor Pattern
The Gang of Four defines the Visitor as:

Represent an operation to be performed on elements of an object structure. Visitor lets you define a new operation without changing the classes of the elements on which it operates.

The nature of the Visitor makes it an ideal pattern to plug into public APIs thus allowing its clients to perform operations on a class using a "visiting" class without having to modify the source.

