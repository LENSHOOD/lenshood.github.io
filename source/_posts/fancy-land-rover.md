---
title: Attempt of Fancy Land Rover
date: 2025-12-07 12:02:18
tags:
- unsupervised learning
- DQN
---

There was a very interesting lab in the course of [Unsupervised Learning, Recommenders, Reinforcement Learning](https://www.coursera.org/learn/unsupervised-learning-recommenders-reinforcement-learning?specialization=machine-learning-introduction) taught by Prof. Andrew Ng, which train a machine learning model to make sure the lunar lander land in a pre-defined range of area. 

The entire lab is based on the [Gymnasium](https://gymnasium.farama.org/), which provides experiential environments for reinforcement learning. After finished the lab, I noticed the Gymnasium is quite flexible and easy to extend, making it possible to play the env in some fancy ways beyond the default configuration.

 <!-- more -->



## Inverted Hover Helicopter & DQN Approach





## Lunar Lander Hover

To make the LunarLander hover in the air instead of landing, basically we'll need to redefine the reward model.

According to the original LunarLander, the target is to land on the ground with 2 legs, and the landing position should be between the two flags:

![](https://gymnasium.farama.org/_images/lunar_lander.gif)

Per the [documentation](https://gymnasium.farama.org/environments/box2d/lunar_lander/) of LunarLander-V3, the reward model is:

```
For each step, the reward:

is increased/decreased the closer/further the lander is to the landing pad.
is increased/decreased the slower/faster the lander is moving.
is decreased the more the lander is tilted (angle not horizontal).
is increased by 10 points for each leg that is in contact with the ground.
is decreased by 0.03 points each frame a side engine is firing.
is decreased by 0.3 points each frame the main engine is firing.

The episode receive an additional reward of -100 or +100 points for crashing or landing safely respectively.
An episode is considered a solution if it scores at least 200 points.
```

Comparing the landing reward, seems hover can be way more easier, since we only need to limit the position vertically and horizontally, no need to concern of fuel and landing posture.

### Extend the original LunarLander

To adjust the original LunarLander meeting our new target: hover, the reward model needs to be updated. And benefit by the open sourced code of gymnasium, we can directly extend the LunarLander by override a few methods.

```python
class FancyLunarLander(LunarLander):
    def __init__(self, x_range, y_range, **kwargs):
        super().__init__(**kwargs)
        self.max_step_reward = 1.0
        self.max_angle_error = 1.0
        self.safe_x_range = x_range

        self.target_y = 1.0
        self.max_height_error = y_range

    def update_range(self, x_range, y_range):
        self.safe_x_range = x_range
        self.max_height_error = y_range
    
    def step(self, action):
        obs, reward, terminated, truncated, info = super().step(action)
        x, y = obs[:2]
        
        # horizontal
        if abs(x) <= self.safe_x_range:
            zone_reward = 0.2 * (1 - abs(x) / self.safe_x_range)
        else:
            zone_reward = -2.0 * (abs(x) - self.safe_x_range)

        # vertical
        height_err = abs(y - self.target_y)
        if height_err < self.max_height_error:
            height_reward = 0.5 * (1 - height_err / self.max_height_error)
        else:
            height_reward = -1.0 * (height_err - self.max_height_error)

        # composed rewards
        reward = zone_reward + height_reward

        return obs, reward, terminated, truncated, info
```

In the override method `step()` , we at first calling the super method to get necessary outputs such as the observations, terminated, truncated and information, which are no need to change at all and will be returned by our override method.

The only change introduced is the "reward", because in our "FancyLunarLander" we don't want to follow the original reward model, which is designed to land in a limited area. Our new goal is to hover, so what we need is let the LunarLander stick in a 2D area, which we can define the vertical and horizontal positions.

According to the override method, obviously, we limit the 2D area to a square that coordinates are `x = [-x_range, x_range]`, `y = [1 - y_range, 1 + y_range]`. The less the x_range / y_range are, the smaller the square is.

Besides, the reward model needs to be redesigned, which as per the code, if the current position of the LunarLander is in the square, it increases the reward by adding positive values (zone_reward and height_reward) to the reward variable, on the contrary, if the current position is out of the square, the reward will be decreased due to negative values to be added.



## Lunar Lander Inverted Hover

The previous chapter proves it's relatively easy to stay a hover status if appropriate rewards are provided. So can we have an inverted hover if we set appropriate rewards as well?

The answer is no, inverted hovering wouldn't be that easy to build like a plain hover. Two obstacles preventing us from achieving that:

1. Standard lunar lander can only produce positive thrust. 

   To make lunar lander hovering, all we need to do is adjusting the main power to produce positive thrust combining with fine adjustments of left/right engines. But it's impossible for having positive thrust only if the lunar lander is in the inverted status,  since positive thrust would just accelerate the lunar lander to crash.

2. Inverted position is not a common and easily achievable posture

   In Gymnasium, the initial status of lunar lander, is in upright position with random velocity/angular velocity, it's difficult & inefficient to set several rewards toward inverted hover state and letting reenforcement learning model to discover the correct movements by itself.

To overcome the above 2 obstacles, first we need change the lunar lander source code to have negative thrust, then we are going to use the approaches from the paper, which is “*Apprenticeship Learning for Target Trajectory*”, to learn from expert trajectory.

### Negative thrust

In Gymnasium, there are two envs of lunar lander, discrete or continuous. At this time we are going to use the continuous env which allows us passing `Box(-1, +1, (2,), dtype=np.float32)` as the action for more precise control. For the lunar lander, the main engine will be turned off completely if `main < 0` and the throttle scales affinely from 50% to 100% for `0 <= main <= 1`

What we need to modify is to unlock the limitation of no power if `main < 0`. Instead, we would want the main engine outputs negative thrust if `main < 0`. 

The version of Gymnasium we are using is v1.2.3, locate to the code we can find the limitation of main power at [here](https://github.com/Farama-Foundation/Gymnasium/blob/43965e15c2424a2b6955c79e774b0810457fd5be/gymnasium/envs/box2d/lunar_lander.py#L535):

``` python
f self.continuous:
    m_power = (np.clip(action[0], 0.0, 1.0) + 1.0) * 0.5  # 0.5..1.0
    assert m_power >= 0.5 and m_power <= 1.0
else:
    m_power = 1.0
```

As we'll use the "continuous" mode so we only need to change the `clip` and `assert` to the following:

``` python
m_power = (np.clip(action[0], -1.0, 1.0))
```

With the minor change, we create a class named `BidirectionalLunarLander`, extends from LunarLander and override the entire `step()` method containing the change as well. (Full code see: )

Now we have negative thrust, let's move on to the next step.

### Apprenticeship Learning for Target Trajectory

Considering the difficulty of pushing the RF model to randomly learn a way to flip, it's better to positively find a way that probably drives the lunar lander flip to inverted posture. Once it goes into inverted, then we try to make it stable in the position.

Based on the `BidirectionalLunarLander`, with several attempts, I created a `naive_inverted_controller` that can produce actions to make the lunar lander flip in most cases. It has nothing to do with "learning", it also isn't an optimal controller, it's just a  "naive" controller:

```python
def navie_inverted_controller(obs, phase, sign):
    x = obs[0]
    y = obs[1]
    vx = obs[2]
    vy = obs[3]
    theta = obs[4]
    omega = obs[5]

    # angle wrap to [-pi, pi]
    theta_wrapped = np.arctan2(np.sin(theta), np.cos(theta))
    theta_abs = abs(theta_wrapped)

    side = 0
    main = 0
    if phase == 1:
        if y > 1.4:
            return np.array([0, 0], dtype=np.float32), phase
            
        if theta_abs < 1.9:
            side = -0.6 * sign
            main = 0.8
        elif theta_abs > 1.9 and abs(omega) > 0.1 :
            side = 1 * sign
            main = -1
        else:
            phase = 2

    if phase == 2:
        if y > 1.4:
            return np.array([0, 0], dtype=np.float32), phase

        main = -0.6
        if vx > 0.1:
            side = 0.5
        elif vx < -0.1:
            side = -0.5
        else:
            side = 0

    return np.array([
        np.clip(main, -1.0, 1.0),
        np.clip(side, -1.0, 1.0)
    ], dtype=np.float32), phase
```

Many hard coded statics are in the implementation, but don't worry, it works on our `BidirectionalLunarLander` and can be used to generate training data later.

Once we can confidently drive the lunar lander to flip, next is how to make it stable in the inverted position without crash. Obviously our naive controller is only good at flip, it cannot fine control the lunar lander towards stabilization. But with the experience of the previous chapter - lunar lander hover - it's straightforward to stable the lunar lander by using reenforcement learning.

This time as we are using the `continuous` control model, DQN might not be a good choice. SAC will be better.
