// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

import * as G6 from '@antv/g6';
import { nodeTextHandle } from './topology';

interface Size {
    width: number;
    height: number;
}
/**
 * 
 * @param textLng 
 * @param height LabelHeight
 * @param paddingLR paddingLeft与paadingRight
 * @param spacingLeft 文本与图片间的间距
 * @param imgWidth 
 */
 export const setLabelSize = ( textLng: number, 
    height: number,
    paddingLR: number,
    spacingLeft?: number, 
    imgWidth?: number
): Size => {
    return {
        width: textLng + (spacingLeft || 0) + (imgWidth || 0) + 2 * paddingLR,
        height
    }
} 
interface Point {
    x: number;
    y: number;
}
export const isShowLabel = (path: any, labelWidth: number): boolean => {
    const startPoint: Point = path.getPoint(0)
    const endPoint: Point = path.getPoint(1)
    const lineLng: number = G6.Util.distance(startPoint, endPoint)
    return !(labelWidth > lineLng ? false : true)
}
// custom the edge with an extra rect
G6.registerEdge('service-edge', {
    lebelPosition: 'center',
    labelAutoRotate: true,
    afterDraw: (cfg: any, group: any) => {
        // 获取路径中点坐标
        const edge = group.get('children')[0]
        const midPoint = edge.getPoint(0.5)
        const flowLabel = group.addGroup({ id: 'flowLabel' })
        
        if (midPoint.x && cfg.service) {
            let text = nodeTextHandle(cfg.service, 10);
            const rectColor = cfg.rectColor || '#d9d9d9';
            const [textWidth] = G6.Util.getTextSize(text, 12)  // flow.attr('fontSize')
            const { width: labelWidth, height: labelHeight } = setLabelSize(textWidth, 16, 5, 5, 10)
            let flow = flowLabel.addShape('text', {
                attrs: {
                    x: midPoint.x + 6,
                    y: midPoint.y + 6,
                    fill: '#FFFFFF',
                    textAlign: 'center',
                    text: text
                },
                name: 'service-node-text',
                zIndex: 1000
            })
            let pointShape = flowLabel.addShape('circle', {
                name: 'circle-shape',
                attrs: {
                    x: midPoint.x - labelWidth / 2 + 8,
                    y: midPoint.y,
                    r: 3.5,
                    fill: '#dcdcdc'
                },
                zIndex: 1000
            })
            let labelBg = flowLabel.addShape('rect', {
                attrs: {
                    // label 在线中点
                    x: midPoint.x - labelWidth / 2,
                    y: midPoint.y - labelHeight / 2,
                    width: labelWidth,
                    height: labelHeight,
                    radius: [4],
                    fill: '#141619',
                    stroke: rectColor
                },
                id: 'service-node-rect',
                name: 'service-node-rect',
                draggable: true,
                zIndex: 10
            })      
            const offsetStyle = G6.Util.getLabelPosition(edge, 0.5, 0, 0, true);
            labelBg.rotateAtPoint(midPoint.x, midPoint.y, offsetStyle.rotate);
            pointShape.rotateAtPoint(midPoint.x, midPoint.y, offsetStyle.rotate);
            flow.rotateAtPoint(midPoint.x, midPoint.y, offsetStyle.rotate);
            flowLabel.sort()
            if (isShowLabel(edge, labelWidth)) {
                // 慎用destroy()
                group.removeChild(group.findById('flowLabel'))
            }
        }
    },
    // * 为了获取到midPoint
    update: undefined,
}, 'line');

G6.registerEdge('service-edge2', {
    lebelPosition: 'center',
    labelAutoRotate: true,
    afterDraw: (cfg: any, group: any) => {
        // 获取路径中点坐标
        const edge = group.get('children')[0]
        const midPoint = edge.getPoint(0.5)
        const flowLabel = group.addGroup({ id: 'flowLabel' })
        
        if (midPoint.x && cfg.service) {
            let text = nodeTextHandle(cfg.service, 10);
            const rectColor = cfg.rectColor || '#d9d9d9';
            const [textWidth] = G6.Util.getTextSize(text, 12)  // flow.attr('fontSize')
            const { width: labelWidth, height: labelHeight } = setLabelSize(textWidth, 16, 5, 5, 10)
            let flow = flowLabel.addShape('text', {
                attrs: {
                    x: midPoint.x + 6,
                    y: midPoint.y + 6,
                    fill: '#FFFFFF',
                    textAlign: 'center',
                    text: text
                },
                name: 'service-node-text',
                zIndex: 1000
            })
            let pointShape = flowLabel.addShape('circle', {
                name: 'circle-shape',
                attrs: {
                    x: midPoint.x - labelWidth / 2 + 8,
                    y: midPoint.y,
                    r: 3.5,
                    fill: '#dcdcdc'
                },
                zIndex: 1000
            })
            let rectShape = flowLabel.addShape('rect', {
                attrs: {
                    // label 在线中点
                    x: midPoint.x - labelWidth / 2,
                    y: midPoint.y - labelHeight / 2,
                    width: labelWidth,
                    height: labelHeight,
                    radius: [4],
                    fill: '#141619',
                    stroke: rectColor
                },
                id: 'service-node-rect',
                name: 'service-node-rect',
                draggable: true,
                zIndex: 10
            })      
            const offsetStyle = G6.Util.getLabelPosition(edge, 0.5, 0, 0, true);
            rectShape.rotateAtPoint(midPoint.x, midPoint.y, offsetStyle.rotate);
            pointShape.rotateAtPoint(midPoint.x, midPoint.y, offsetStyle.rotate);
            flow.rotateAtPoint(midPoint.x, midPoint.y, offsetStyle.rotate);
            flowLabel.sort()
            if (isShowLabel(edge, labelWidth)) {
                // 慎用destroy()
                group.removeChild(group.findById('flowLabel'))
            }
        }
    },
    // * 为了获取到midPoint
    update: undefined,
}, 'quadratic');
