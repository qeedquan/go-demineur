#! /usr/bin/python
# -*- coding: utf-8 -*-

from pygame import *
from random import sample

Tab_w,Tab_h = 40,30
Screen = display.set_mode((Tab_w*30,Tab_h*30))
Screen_rect = Screen.get_rect()

Nb_mines = (Tab_w*Tab_h)/8

master = image.load('master.png')
Img = []
for x in range(16):
    Img.append(Surface((30,30)))
    Img[-1].blit(master,(0,0),(x*30,0,30,30))
mine_bleu,bleu,croix,flip1,flip2,mine_rouge,noir = Img[9:]

win = image.load('win.png')
lose = image.load('lose.png')

font.init()
hit_space = font.Font(font.get_default_font(),20).render('HIT SPACE TO RESTART',1,(100,0,0))
hit_space_rect = hit_space.get_rect()

class Case(Rect):
    ''
    def __init__(self,rect):
        Rect.__init__(self,rect)
        self.valeur = 0
        self.warn = False
        self.vue = False

def flip(temp):
    display.update([Screen.blit(flip1,Tab[x][y]) for x,y in temp])
    time.wait(50)
    display.update([Screen.blit(noir,Tab[x][y]) for x,y in temp])
    time.wait(50)
    display.update([Screen.blit(flip2,Tab[x][y]) for x,y in temp])
    time.wait(100)
    display.update([Screen.blit(Img[Tab[x][y].valeur],Tab[x][y]) for x,y in temp])

time.set_timer(USEREVENT,1000)

while True:
    base_time = time.get_ticks()
    Reste_cases = Tab_w*Tab_h - Nb_mines
    Tab = [[Case(Screen.blit(bleu,(x*30,y*30))) for y in range(Tab_h)]for x in range(Tab_w)]
    Mines = [(i%Tab_w,i/Tab_w) for i in sample(range(Tab_w*Tab_h),Nb_mines)]
    for X,Y in Mines:
        Tab[X][Y].valeur = 9
        for x,y in (-1,-1),(0,-1),(1,-1),(-1,0),(1,0),(-1,1),(0,1),(1,1):
            if 0<=X+x<Tab_w and 0<=Y+y<Tab_h and Tab[X+x][Y+y].valeur < 9:
                Tab[X+x][Y+y].valeur += 1
    Gagnant = True
    display.update()
    while Reste_cases:
        ev = event.wait()
        
        if ev.type == USEREVENT:
            laps = (time.get_ticks()-base_time)/1000
            sec = str(laps%60).zfill(2)
            min = str(laps/60%60).zfill(2)
            heure = str(laps/3600%24).zfill(2)
            display.set_caption('%s:%s:%s    Reste %s cases'%(heure,min,sec,Reste_cases))
        
        elif ev.type == MOUSEBUTTONUP:
            x,y = ev.pos[0]/30%Tab_w,ev.pos[1]/30%Tab_h
            if not Tab[x][y].vue:
                if ev.button == 1 and not Tab[x][y].warn:
                    Tab[x][y].vue = True; Reste_cases -= 1
                    temp = [(x,y)]
                    for X,Y in temp:
                        if not Tab[X][Y].valeur:
                            for ofsx,ofsy in (-1,-1),(0,-1),(1,-1),(-1,0),(1,0),(-1,1),(0,1),(1,1):
                                if 0<=X+ofsx<Tab_w and 0<=Y+ofsy<Tab_h and not (Tab[X+ofsx][Y+ofsy].vue or Tab[X+ofsx][Y+ofsy].warn):
                                    Tab[X+ofsx][Y+ofsy].vue = True; Reste_cases -= 1
                                    temp.append((X+ofsx,Y+ofsy))
                    flip(temp)
                    if Tab[x][y].valeur == 9:
                        Gagnant = False
                        break
                elif ev.button == 3:
                    Tab[x][y].warn = not Tab[x][y].warn
                    display.update(Screen.blit(croix if Tab[x][y].warn else bleu,Tab[x][y]))
        
        elif ev.type == QUIT: exit()
        
    if not Gagnant:
        for _ in range(5):
            display.update(Screen.blit(mine_bleu if _&1 else mine_rouge,Tab[x][y]))
            time.wait(500)
    else:
        flip(Mines)
        time.wait(500)
        
    losewin = win if Gagnant else lose
    rect = losewin.get_rect()
    rect.center = Screen_rect.center
    Screen.blit(losewin,rect)
    hit_space_rect.midbottom = Screen_rect.centerx,Screen_rect.bottom-28
    draw.rect(Screen,(0,0,100),Screen.fill(-1,(0,hit_space_rect.top-10,Screen_rect.w,40)),1)
    Screen.blit(hit_space,hit_space_rect)
    display.flip()
    while not key.get_pressed()[K_SPACE]:
        if event.wait().type == QUIT: exit()
