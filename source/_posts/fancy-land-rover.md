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



## Inverted Hover Helicopter & DQN approach





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



## Lunar Lander Inverted Hover

