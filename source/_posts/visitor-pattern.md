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

### Implement Visitor Pattern
So now we know that in the traditional we use Parent-Children inheritance to do polymorphic, but there's a problem that if we happend to need add a new behavior into parent, then we have to spend many time to deal with a disaster, which is add that behavior implementation to every single child.

Just like the above instance, assume that we refactor the code to meet the OO principle:
``` java
public interface Shape {
	double area(Object shape);
}

public class Square implements Shape {
	public Point topLeft;
	public double side;
	
	public double area(Object shape) {
		return side * side;
	}
}

public class Rectangle implements Shape {
	public Point topLeft;
	public double height;
	public double width;
	
	public double area(Object shape) {
		return height * width;
	}
}

public class Circle implements Shape {
	public Point center;
	public double radius;
	public final double PI = 3.141592653589793;
	
	public double area() {
		return PI * radius * radius;
	}
}

public class Geometry {
	public double area(Shape shape) {
		return shape.area();
	}
}
```
Woo, simple and elegant!

Unfortunately, uncle Bob want us to add a perimeter() to Geometry, in this time, only if we add such behavior to each Shape can solve that problem, so let's do it!

However, we sadly find that all the Shape code have already deployed to production, so we cannot just modify Shape to meet the new requirment because we may introducing potential risks to the old code.

Seems we ended in a deadlock, it's time to introduing Visitor Pattern:
We leave aside how to build a vistor, just see the code as follow:
``` java
public interface Visitor {
	void visit(Square square);
	void visit(Rectangle rectangle);
	void visit(Circle circle);
}

public class PerimeterVisitor implements Visitor {
	private double perimeter;
	
	public void visit(Square square) {
		perimeter = square.side * 4;
	}
	
	public void visit(Rectangle rectangle) {
		perimeter = rectangle.height * 2 + rectangle.width * 2;
	}
	
	public void visit(Circle circle) {
		perimeter = circle.radius * 2 * circle.PI;
	}
}

public class AreaVisitor implements Visitor {
	private double area;
	
	public void visit(Square square) {
		area = square.side * square.side;
	}
	
	public void visit(Rectangle rectangle) {
		area = rectangle.height * rectangle.width;
	}
	
	public void visit(Circle circle) {
		area = circle.PI * circle.radius * circle.radius;
	}
}

public interface Shape {
	void accept(Visitor v);
}

public class Square implements Shape {
	public Point topLeft;
	public double side;
	
	public void accept(Visitor v) {
		v.visit(this);
	}
}

public class Rectangle implements Shape {
	public Point topLeft;
	public double height;
	public double width;
	
	public void accept(Visitor v) {
		v.visit(this);
	}
}

public class Circle implements Shape {
	public Point center;
	public double radius;
	public final double PI = 3.141592653589793;
	
	public void accept(Visitor v) {
		v.visit(this);
	}
}

public class Geometry {
	public double area(Shape shape) {
		AreaVisitor areaVisitor = new AreaVisitor();
		shape.accept(areaVisitor);
		return areaVisitor.area;
	}
	
	public perimeter(Shape shape) {
		PerimeterVisitor perimeterVisitor = new PerimeterVisitor();
		shape.accept(perimeterVisitor);
		return perimeterVisitor.perimeter;
	}
}
```
Look like we create something new named Visitor, and created  two Visitor implementations. Then at Shape, we added a new method `accept(Visitor v)`, which takes a Visitor as parmeter. 

As we can see, unlike the inheritance plan, this new version of code remove behavior inside the Shape implementations, otherwise, move it to the implementation of Visitor -- AreaVisitor and PerimeterVisitor.

That's the core concept of Visitor Pattern: separate data structure and behavior to make it mor e flexible to modify.

Considering deal with different class type, we put three `visit()` method into the Visitor related to three type of Shape. And put `accept(Visitor v)`  into each Shape, let Shape itself to choose the right `visit()`, rather than using a bunch of `instance of` to distinguish different type. This is what we called: double dispatch.

### Abstraction of Visitor
We alreay known how to use vistor pattern to solve the problem of behavior change. Now let's conclude and describe what exactly is vistor.

From the code at above, we can see two parts: visitor, concrete vistor, shape, concrete shap. For more generality, now we call the shape as element, then the concrete shape such as square, circle, we can call them concrete element. 

After that, we can abstract the UML of visitor:

Hence, we can get:
1. Visitor provide several visit(Element e) method to meet every type of element. visit(Element e) take a Element as parameter, and get useful information from that element to do some job.
2. Element provide accept(Vistor v) method to "accept" a visitor, then do the standard operation: `v.visit(this);`, throught this, element can halp visitor to dynamicly run the right visit method, to achieve double dispatch.

